package compressed

import (
	"bytes"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"image"
	"image-compressions/config"
	"image-compressions/connector"
	"image-compressions/consts"
	"image-compressions/helper"
	"image-compressions/pkg"
	"image-compressions/request"
	"image/jpeg"
	"os"
	"path"
	"strings"
	"sync"
)

func ConsumerProcessing(logger *logrus.Logger, cfg config.Configurations) {
	var (
		imageFile  string
		outputPath string
		payload    request.DiscordRequest
	)

	rabbitMqCfg := pkg.NewRabbitMQConfig(cfg.RabbitMq.Username, cfg.RabbitMq.Password, cfg.RabbitMq.Port, cfg.RabbitMq.Host)
	msgx, _, _, err := pkg.Consumer(rabbitMqCfg, cfg.RabbitMq.Topic)
	if err != nil {
		logger.Printf("%s-consumer rabbit-mq got error: %v", cfg.Server.Name, err)
		return
	}

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
			case msg, ok := <-msgx:
				if !ok {
					return
				}
				imageFile = string(msg.Body)
				readFl := fmt.Sprintf("%s/%s", cfg.RabbitMq.PathOriginalFile, imageFile)

				if imageFile == "" {
					logger.Printf("failed to read path file : %v, skip process", err)
					payload.Content = fmt.Sprintf("%s-failed to read path file : %v, filename: %s, skip process", cfg.Server.Name, err, imageFile)
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
				newPath := strings.TrimLeft(baseFile, cfg.RabbitMq.PathOriginalFile)
				dirOutput := checkSubDirectory(
					cfg.RabbitMq.SubPathOriginalInvtrypht,
					cfg.RabbitMq.SubPathCompressionInvtrypht,
					cfg.RabbitMq.SubPathOriginalAdjdmgpht,
					cfg.RabbitMq.SubPathCompressionAdjdmgpht,
					prefix,
				)

				outputPath = fmt.Sprintf("%s/%s/%s", cfg.RabbitMq.PathCompressed, dirOutput, newPath)

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
		return helper.ChangeFileExtension(fileImage, consts.JPEG)
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
