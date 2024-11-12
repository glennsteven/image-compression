package compressed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"image"
	"image-compressions/internal/config"
	"image-compressions/internal/connector"
	"image-compressions/internal/request"
	"image-compressions/pkg/helper"
	"image-compressions/pkg/storage"
	"image/jpeg"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"
)

type Consumer struct {
	logger   *logrus.Logger
	delivery <-chan amqp.Delivery
	alerting connector.Alerting
	aws      storage.Storage
}

func NewConsumer(logger *logrus.Logger, delivery <-chan amqp.Delivery, alerting connector.Alerting, aws storage.Storage) *Consumer {
	return &Consumer{
		logger:   logger,
		delivery: delivery,
		alerting: alerting,
		aws:      aws,
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
				break
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

	if req.FileName == "" {
		c.logAndNotifyError(fmt.Sprintf("cannot found image: %v, server config: %s, skip process", req.FileName, cfg.Server.Name))
		return
	}

	awsCtx := context.Background()
	byteImage, err := c.aws.Get(awsCtx, cfg.Aws.Bucket, req.FileName)
	if err != nil {
		c.logger.Printf("func aws.Get failed fetch byte from aws: %v, [filename: %s]", err, req.FileName)
		c.logAndNotifyError(fmt.Sprintf("%s-func aws.Get failed fetch byte from aws : %v, filename: %s", cfg.Server.Name, err, req.FileName))
		return
	}

	fileImage, isConv, err := helper.ToJpeg(byteImage)
	if err != nil {
		c.logger.Printf("convert image got error %v || [filename: %s]", err, req.FileName)
		c.logAndNotifyError(fmt.Sprintf("%s-convert image got error %v, filename %s", cfg.Server.Name, err, req.FileName))
		return
	}

	img, _, err := image.Decode(bytes.NewReader(fileImage))
	if err != nil {
		c.logger.Printf("Error decoding the image: %v", err)
		c.logAndNotifyError(fmt.Sprintf("%s-Error decoding the image: %v || [filename: %s]", cfg.Server.Name, err, req.FileName))
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

	// Create an in-memory buffer to store the JPEG data
	var buf bytes.Buffer

	// Encode the image as JPEG with compression options
	jpegOptions := jpeg.Options{Quality: cfg.ImageSetting.Quality}

	err = jpeg.Encode(&buf, img, &jpegOptions)
	if err != nil {
		c.logger.Printf("Error encoding the image: %v", err)
		c.logAndNotifyError(fmt.Sprintf("%s-Error encoding the image: %v, %s", cfg.Server.Name, err, req.FileName))
		msg.Nack(false, true)
		return
	}

	uploadAws := fmt.Sprintf("%s/%s", dirOutput, newPath)
	err = c.aws.Put(awsCtx, cfg.Aws.Bucket, uploadAws, buf.Bytes(), req.MimeType)
	if err != nil {
		c.logger.Printf("func aws.Put failed upload file: %v", err)
		c.logAndNotifyError(fmt.Sprintf("%s-Error upload file to aws: %v, %s", cfg.Server.Name, err, req.FileName))
		return
	}

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
