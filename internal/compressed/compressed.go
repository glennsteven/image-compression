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
	_ "image/png"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Consumer struct {
	logger   *logrus.Logger
	alerting connector.Alerting
	aws      storage.Storage
	wg       sync.WaitGroup
	cfg      *config.Configurations
}

func NewConsumer(logger *logrus.Logger, alerting connector.Alerting, aws storage.Storage, cfg *config.Configurations) *Consumer {
	return &Consumer{
		logger:   logger,
		alerting: alerting,
		aws:      aws,
		cfg:      cfg,
	}
}

func (c *Consumer) Listen(ctx context.Context, delivery <-chan amqp.Delivery) {
	poolSize := c.cfg.Rabbitmq.PoolSize
	c.logger.Printf(" [*] Waiting for messages. Start %d worker...", poolSize)

	for i := 0; i < poolSize; i++ {
		c.wg.Add(1)
		go c.worker(i+1, delivery)
	}

	// Wait for the context to complete (e.g., when receiving a shutdown signal)
	<-ctx.Done()
	c.logger.Println("Shutdown signal received, waiting for all workers to finish...")

	// Wait for all running workers to complete their tasks
	c.wg.Wait()
	c.logger.Println("All workers have quit.")
}

// worker is a goroutine that will run continuously processing messages.
func (c *Consumer) worker(id int, delivery <-chan amqp.Delivery) {
	defer c.wg.Done()
	logger := c.logger.WithField("worker_id", id)
	logger.Println("Worker starts")

	for msg := range delivery {
		c.processMessage(logger, msg)
	}

	logger.Println("Workers quit.")
}

func (c *Consumer) processMessage(logger *logrus.Entry, msg amqp.Delivery) {
	var req request.ConsumerRequest
	if err := json.Unmarshal(msg.Body, &req); err != nil {
		logger.Errorf("Failed unmarshal request: %v. The message will be discarded.", err)
		// This message is corrupted, do not requeue it. Nack(requeue=false)
		_ = msg.Nack(false, false)
		return
	}

	// Add context to logger for better tracking
	logger = logger.WithField("filename", req.FileName)

	if req.FileName == "" {
		logger.Warn("Nama file kosong, pesan dilewati.")
		// Assume success (Ack) so that the message is not reprocessed.
		_ = msg.Ack(false)
		return
	}

	// Breaking logic into smaller functions
	err := c.compressAndUpload(logger, &req)
	if err != nil {
		logger.Errorf("Failed to process image: %v", err)
		c.logAndNotifyError(fmt.Sprintf("%s - Failed to process file %s: %v", c.cfg.Server.Name, req.FileName, err))
		// An error occurred, returning the message to the queue for retry (Nack, requeue=true)
		_ = msg.Nack(false, true)
		return
	}

	logger.Info("Image successfully compressed and uploaded")

	// All processes are successful, send acknowledgement (Ack)
	_ = msg.Ack(false)
}

func (c *Consumer) compressAndUpload(logger *logrus.Entry, req *request.ConsumerRequest) error {
	//ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	//defer cancel()

	// TODO: If you have AWS Account, please uncomment this code
	// 1. Take a picture from S3
	//byteImage, err := c.aws.Get(ctx, c.cfg.Aws.Bucket, req.FileName)
	//if err != nil {
	//	return fmt.Errorf("failed to retrieve from S3: %w", err)
	//}

	pathFile := fmt.Sprintf("***/original/%s", req.FileName)
	byteImage, err := os.ReadFile(pathFile)
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}

	logger.Info(fmt.Sprintf("Successfully opened local file: %s", pathFile))

	// 2. Convert and Decode
	fileImage, isConv, err := helper.ToJpeg(byteImage)
	if err != nil {
		return fmt.Errorf("failed to convert to Jpeg: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(fileImage))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// 3. Determine the output path
	outputPath := c.buildOutputPath(req.FileName, isConv)

	// 4. Compressed (Encode) Image
	var buf bytes.Buffer
	jpegOptions := jpeg.Options{Quality: c.cfg.ImageSetting.Quality}
	if err = jpeg.Encode(&buf, img, &jpegOptions); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	// TODO: If you have AWS Account, please uncomment this code
	// 5. Upload to S3
	//uploadPath := strings.TrimPrefix(outputPath, c.cfg.ImageSetting.PathCompressed+"/")
	//if err = c.aws.Put(ctx, c.cfg.Aws.Bucket, uploadPath, buf.Bytes(), req.MimeType); err != nil {
	//	return fmt.Errorf("failed to upload to S3: %w", err)
	//}

	fullPath := filepath.Join(outputPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}

	// write file
	if err := os.WriteFile(fullPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write local file: %w", err)
	}

	return nil
}

// buildOutputPath is a helper for building output file paths.
func (c *Consumer) buildOutputPath(fileName string, isConv bool) string {
	//outputImage := determineOutputImage(fileName, isConv)
	//baseFile := path.Base(outputImage)
	//prefix := strings.TrimSuffix(outputImage, baseFile)
	//newPath := strings.TrimPrefix(baseFile, c.cfg.ImageSetting.PathOriginalFile)
	//
	//dirOutput := checkSubDirectory(
	//	c.cfg.ImageSetting.SubPathOriginalInvtrypht,
	//	c.cfg.ImageSetting.SubPathCompressionInvtrypht,
	//	c.cfg.ImageSetting.SubPathOriginalAdjdmgpht,
	//	c.cfg.ImageSetting.SubPathCompressionAdjdmgpht,
	//	prefix,
	//)

	return fmt.Sprintf("****/result/%s", fileName)
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
