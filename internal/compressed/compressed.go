package compressed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"image"
	"image-compressions/internal/config"
	"image-compressions/internal/connector"
	"image-compressions/internal/request"
	"image-compressions/pkg/helper"
	"image/jpeg"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"
)

var backoffSchedule = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	10 * time.Second,
}

type Consumer struct {
	logger   *logrus.Logger
	delivery <-chan amqp.Delivery
	alerting connector.Alerting
}

func NewConsumer(logger *logrus.Logger, delivery <-chan amqp.Delivery, alerting connector.Alerting) *Consumer {
	return &Consumer{
		logger:   logger,
		delivery: delivery,
		alerting: alerting,
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
				continue
			}
			go c.consume(msg, cfg)
		}
	}
}

func (c *Consumer) consume(msg amqp.Delivery, cfg *config.Configurations) {
	var (
		outputPath string
		req        request.ConsumerRequest
	)

	err := json.Unmarshal(msg.Body, &req)
	if err != nil {
		c.logger.Printf("cannot unmarshal request : %q", err)
		return
	}

	readFile := fmt.Sprintf("%s/%s", cfg.ImageSetting.PathOriginalFile, req.FileName)

	if req.FileName == "" {
		c.logAndNotifyError(fmt.Sprintf("cannot found image : %v, server config: %s, skip process", req.FileName, cfg.Server.Name))
		return
	}

	fileBytes, err := c.readFileRetries(readFile)
	if err != nil {
		c.logger.Printf("failed to read path file : %v", err)
		c.logAndNotifyError(fmt.Sprintf("%s-failed to read path file : %v, filename: %s", cfg.Server.Name, err, req.FileName))
		return
	}

	if len(fileBytes) == 0 {
		c.logger.Printf("%s-Skip processing file: %s, %v bytes", cfg.Server.Name, req.FileName, len(fileBytes))
		c.logAndNotifyError(fmt.Sprintf("%s-Skip processing file: %s, %v bytes", cfg.Server.Name, req.FileName, len(fileBytes)))
		return
	}

	fileImage, isConv, err := helper.ToJpeg(fileBytes)
	if err != nil {
		c.logger.Printf("convert image got error %v", err)
		c.logAndNotifyError(fmt.Sprintf("%s-convert image got error %v, filename %s", cfg.Server.Name, err, req.FileName))
		msg.Nack(false, false)
		return
	}

	c.logger.Println("Image conversion successfully!")

	img, _, err := image.Decode(bytes.NewReader(fileImage))
	if err != nil {
		c.logger.Printf("Error decoding the image: %v", err)
		c.logAndNotifyError(fmt.Sprintf("%s-Error decoding the image: %v", cfg.Server.Name, err))
		msg.Nack(false, false)
		return
	}

	outputImage := determineOutputImage(req.FileName, isConv)

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
		c.logAndNotifyError(fmt.Sprintf("%s-Error creating the output image: %v, %s", cfg.Server.Name, err, baseFile))
		msg.Nack(false, true)
		return
	}

	//Todo: Handle unique filename if needed to ensure not replace same filename

	// Encode the image as JPEG with compression options
	jpegOptions := jpeg.Options{Quality: cfg.ImageSetting.Quality}

	err = jpeg.Encode(output, img, &jpegOptions)
	if err != nil {
		c.logger.Printf("Error encoding the image: %v", err)
		c.logAndNotifyError(fmt.Sprintf("%s-Error encoding the image: %v,%v", cfg.Server.Name, err, output))
		msg.Nack(false, true)
		return
	}

	output.Close()

	logrus.Infof("%s-Image compressed and saved in %s\n", cfg.Server.Name, outputPath)
	logrus.WithFields(logrus.Fields{"server": cfg.Server.Name}).Info("Successfully uploaded file!")
}

func (c *Consumer) logAndNotifyError(text string) {
	c.logger.Println("[logAndNotifyError] ", text)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	errConnector := c.alerting.SendAlert(ctx, text)
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

func (c *Consumer) readFileRetries(url string) (fileBytes []byte, err error) {
	for _, backoff := range backoffSchedule {
		fileBytes, err = os.ReadFile(url)

		if err == nil {
			break
		}

		c.logger.Printf("Request error: %+v\n", err)
		c.logger.Printf("Retrying in %v\n", backoff)
		time.Sleep(backoff)
	}

	// All retries failed
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}
