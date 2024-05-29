package compressed

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"image"
	"image-compressions/internal/config"
	"image-compressions/internal/connector"
	"image-compressions/pkg/helper"
	"image-compressions/request"
	"image/jpeg"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
)

type Consumer struct {
	logger   *logrus.Logger
	delivery <-chan amqp.Delivery
}

func NewConsumer(logger *logrus.Logger, delivery <-chan amqp.Delivery) *Consumer {
	return &Consumer{
		logger:   logger,
		delivery: delivery,
	}
}

func (c *Consumer) Listen(cfg *config.Configurations) {
	c.logger.Printf(" [*] Waiting for messages. To exit press CTRL+C \n")
	shutDownListener := make(chan os.Signal, 1)
	signal.Notify(shutDownListener, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case sig := <-shutDownListener:
			c.logger.Printf("shutdown requested signal: %s \n", sig.String())
			return
		case msg, ok := <-c.delivery:
			if !ok {
				c.logger.Printf("consumer is closed")
				return
			}
			go c.consume(msg, cfg)
		}
	}
}

func (c *Consumer) consume(msg amqp.Delivery, cfg *config.Configurations) {
	var (
		imageFile  string
		outputPath string
		payload    request.DiscordRequest
	)

	imageFile = string(msg.Body)
	readFl := fmt.Sprintf("%s/%s", cfg.ImageSetting.PathOriginalFile, imageFile)

	if imageFile == "" {
		payload.Content = fmt.Sprintf("cannot found image : %v, server config: %s, skip process", imageFile, cfg.Server.Name)
		c.logAndNotifyError(cfg.Discord.Url, payload)
		return
	}

	fileBytes, err := os.ReadFile(readFl)
	if err != nil {
		c.logger.Printf("failed to read path file : %v", err)
		payload.Content = fmt.Sprintf("%s-failed to read path file : %v, filename: %s", cfg.Server.Name, err, imageFile)
		c.logAndNotifyError(cfg.Discord.Url, payload)
		msg.Nack(false, true) //requeue
		return
	}

	if len(fileBytes) == 0 {
		c.logger.Printf("%s-Skip processing file: %s, %v bytes", cfg.Server.Name, imageFile, len(fileBytes))
		payload.Content = fmt.Sprintf("%s-Skip processing file: %s, %v bytes", cfg.Server.Name, imageFile, len(fileBytes))
		c.logAndNotifyError(cfg.Discord.Url, payload)
		return
	}

	fileImage, isConv, err := helper.ToJpeg(fileBytes)
	if err != nil {
		c.logger.Printf("convert image got error %v", err)
		payload.Content = fmt.Sprintf("%s-convert image got error %v, filename %s", cfg.Server.Name, err, imageFile)
		c.logAndNotifyError(cfg.Discord.Url, payload)
		msg.Nack(false, false)
		return
	}

	c.logger.Println("Image conversion successfully!")

	img, _, err := image.Decode(bytes.NewReader(fileImage))
	if err != nil {
		c.logger.Printf("Error decoding the image: %v", err)
		payload.Content = fmt.Sprintf("%s-Error decoding the image: %v", cfg.Server.Name, err)
		c.logAndNotifyError(cfg.Discord.Url, payload)
		msg.Nack(false, false)
		return
	}

	outputImage := determineOutputImage(imageFile, isConv)

	baseFile := path.Base(outputImage)
	prefix := strings.Trim(outputImage, baseFile)
	newPath := strings.TrimLeft(baseFile, cfg.ImageSetting.PathOriginalFile)
	dirOutput := checkSubDirectory(
		cfg.ImageSetting.SubPathOriginalInvtrypht,
		cfg.ImageSetting.SubPathCompressionInvtrypht,
		cfg.ImageSetting.SubPathOriginalAdjdmgpht,
		cfg.ImageSetting.SubPathCompressionAdjdmgpht,
		prefix,
	)

	outputPath = fmt.Sprintf("%s/%s/%s", cfg.ImageSetting.PathCompressed, dirOutput, newPath)

	output, err := os.Create(outputPath)
	if err != nil {
		c.logger.Printf("Error creating the output image: %v", err)
		payload.Content = fmt.Sprintf("%s-Error creating the output image: %v, %s", cfg.Server.Name, err, baseFile)
		c.logAndNotifyError(cfg.Discord.Url, payload)
		msg.Nack(false, true)
		return
	}

	//Todo: Handle unique filename if needed to ensure not replace same filename

	// Encode the image as JPEG with compression options
	jpegOptions := jpeg.Options{Quality: cfg.ImageSetting.Quality}

	err = jpeg.Encode(output, img, &jpegOptions)
	if err != nil {
		c.logger.Printf("Error encoding the image: %v", err)
		payload.Content = fmt.Sprintf("%s-Error encoding the image: %v,%v", cfg.Server.Name, err, output)
		c.logAndNotifyError(cfg.Discord.Url, payload)
		msg.Nack(false, true)
		return
	}

	output.Close()

	logrus.Infof("%s-Image compressed and saved in %s\n", cfg.Server.Name, outputPath)
	logrus.WithFields(logrus.Fields{"server": cfg.Server.Name}).Info("Successfully uploaded file!")

	msg.Ack(true)
}

func (c *Consumer) logAndNotifyError(urlDiscord string, payload request.DiscordRequest) {
	errConnector := connector.LogDiscord(c.logger, urlDiscord, payload)
	if errConnector != nil {
		c.logger.Printf("Error sending request to discord bot: %v", errConnector)
	}
}

func determineOutputImage(fileImage string, isConvert bool) string {
	if isConvert {
		return helper.ChangeFileExtension(fileImage, `jpeg`)
	}
	return fileImage
}

func checkSubDirectory(subOriInv, subCompInv, subOriAdj, subCompAdj, prefix string) string {
	if subOriInv == prefix {
		return subCompInv
	} else if subOriAdj == prefix {
		return subCompAdj
	}
	return prefix
}
