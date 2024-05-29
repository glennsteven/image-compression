package main

import (
	"fmt"
	"os"

	"image-compressions/internal/app/imagecompression"
)

func main() {
	if err := imagecompression.Start(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "could not start app: %v\n", err)
		os.Exit(1)
	}
}
