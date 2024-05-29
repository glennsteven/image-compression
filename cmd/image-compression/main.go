package main

import (
	"fmt"
	"image-compressions/internal/app/imagecompression"
	"os"
)

func main() {
	if err := imagecompression.Start(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "cannot start the application: %v", err)
		os.Exit(1)
	}
}
