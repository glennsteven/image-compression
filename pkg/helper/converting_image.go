package helper

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"path/filepath"
)

func ToJpeg(file []byte) ([]byte, bool, error) {
	jpegImage, conv, err := convertToJPEG(file)
	if err != nil {
		return nil, false, fmt.Errorf("failed to convert image to JPEG: %v", err)
	}

	return jpegImage, conv, nil
}

func convertToJPEG(imageBytes []byte) ([]byte, bool, error) {

	contentType := http.DetectContentType(imageBytes)

	switch contentType {
	case "image/png":
		image, err := decodeImage(bytes.NewReader(imageBytes))
		if err != nil {
			return nil, false, err
		}
		return image, true, nil
	case "image/jpeg":
		return imageBytes, false, nil
	default:
		return nil, false, fmt.Errorf("unsupported content type: %s", contentType)
	}
}

func decodeImage(file io.Reader) ([]byte, error) {
	img, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG image: %v", err)
	}

	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, nil); err != nil {
		return nil, fmt.Errorf("failed to encode image as JPEG: %v", err)
	}

	return buf.Bytes(), nil
}

func ChangeFileExtension(fileName, newExtension string) string {
	base := fileName[:len(fileName)-len(filepath.Ext(fileName))]
	return base + "." + newExtension
}
