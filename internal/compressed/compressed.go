package compressed

import (
	"bytes"
	"context"
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
	"path"
	"strings"
	"sync"
)

func ConsumerProcessing(logger *logrus.Logger, deliveryChan <-chan amqp.Delivery, cfg *config.Configurations) {
	var (
		imageFile  string
		outputPath string
		payload    request.DiscordRequest
	)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		cancel()
		wg.Wait()
	}()

	forever := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(forever)

		for {
			select {
			case msg, ok := <-deliveryChan:
				if !ok {
					return
				}
				imageFile = string(msg.Body)
				readFl := fmt.Sprintf("%s/%s", cfg.ImageSetting.PathOriginalFile, imageFile)

				if imageFile == "" {
					payload.Content = fmt.Sprintf("cannot found image : %v, server config: %s, skip process", imageFile, cfg.Server.Name)
					logAndNotifyError(logger, cfg.Discord.Url, payload)
					continue
				}

				fileBytes, err := os.ReadFile(readFl)
				if err != nil {
					logger.Printf("failed to read path file : %v", err)
					payload.Content = fmt.Sprintf("%s-failed to read path file : %v, filename: %s", cfg.Server.Name, err, imageFile)
					logAndNotifyError(logger, cfg.Discord.Url, payload)
					msg.Nack(false, true) //requeue
					continue
				}

				if len(fileBytes) == 0 {
					logger.Printf("%s-Skip processing file: %s, %v bytes", cfg.Server.Name, imageFile, len(fileBytes))
					payload.Content = fmt.Sprintf("%s-Skip processing file: %s, %v bytes", cfg.Server.Name, imageFile, len(fileBytes))
					logAndNotifyError(logger, cfg.Discord.Url, payload)
					continue
				}

				fileImage, isConv, err := helper.ToJpeg(fileBytes)
				if err != nil {
					logger.Printf("convert image got error %v", err)
					payload.Content = fmt.Sprintf("%s-convert image got error %v, filename %s", cfg.Server.Name, err, imageFile)
					logAndNotifyError(logger, cfg.Discord.Url, payload)
					msg.Ack(false)
					continue
				}

				logger.Println("Image conversion successfully!")

				img, _, err := image.Decode(bytes.NewReader(fileImage))
				if err != nil {
					logger.Printf("Error decoding the image: %v", err)
					payload.Content = fmt.Sprintf("%s-Error decoding the image: %v", cfg.Server.Name, err)
					logAndNotifyError(logger, cfg.Discord.Url, payload)
					msg.Nack(false, true)
					continue
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
					logger.Printf("Error creating the output image: %v", err)
					payload.Content = fmt.Sprintf("%s-Error creating the output image: %v, %s", cfg.Server.Name, err, baseFile)
					logAndNotifyError(logger, cfg.Discord.Url, payload)
					msg.Nack(false, true)
					continue
				}

				//Todo: Handle unique filename if needed to ensure not replace same filename

				// Encode the image as JPEG with compression options
				jpegOptions := jpeg.Options{Quality: cfg.ImageSetting.Quality}

				err = jpeg.Encode(output, img, &jpegOptions)
				if err != nil {
					logger.Printf("Error encoding the image: %v", err)
					payload.Content = fmt.Sprintf("%s-Error encoding the image: %v,%v", cfg.Server.Name, err, output)
					logAndNotifyError(logger, cfg.Discord.Url, payload)
					msg.Nack(false, true)
					continue
				}

				output.Close()

				logrus.Infof("%s-Image compressed and saved in %s\n", cfg.Server.Name, outputPath)
				logrus.WithFields(logrus.Fields{
					"server": cfg.Server.Name,
				}).Info("Successfully uploaded file!")

			case <-ctx.Done():
				return // context canceled, exit goroutine
			}
		}
	}()

	logger.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	<-forever
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

func logAndNotifyError(logger *logrus.Logger, urlDiscord string, payload request.DiscordRequest) {
	errConnector := connector.LogDiscord(logger, urlDiscord, payload)
	if errConnector != nil {
		logger.Printf("Error sending request to Discord bot: %v", errConnector)
	}
}
