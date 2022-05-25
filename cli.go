package main

import (
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/parrotmac/goutil"
)

var (
	inFile      = flag.String("in", "", "Specify input file")
	outFile     = flag.String("out", "", "Specify output file (format is inferred from extension")
	jpegQuality = flag.Int("quality", 100, "Output quality if target is JPEG")
)

func cli() {
	flag.Parse()

	if inFile != nil && *inFile != "" && outFile != nil && *outFile != "" {

		imgFD, err := os.Open(*inFile)
		if err != nil {
			log.Fatalln(err)
		}
		decodedImg, inFmt, err := image.Decode(imgFD)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("Found %s file\n", inFmt)

		fpParts := strings.Split(*outFile, ".")
		extension := fpParts[len(fpParts)-1]

		outFd, err := os.Create(*outFile)
		if err != nil {
			log.Fatalln("could not open output file for writing", err)
		}
		defer outFd.Close()

		var encodeErr error
		switch extension {
		case "jpg", "jpeg":
			encodeErr = JpegDistort(decodedImg, outFd, 5, 1)
		case "png":
			// PNG is lossless, so no options
			encodeErr = png.Encode(outFd, decodedImg)
		case "gif":
			encodeErr = gif.Encode(outFd, decodedImg, &gif.Options{
				NumColors: goutil.Bounded(goutil.UnwrapOr(jpegQuality, 256), 1, 256),
			})
		}
		if encodeErr != nil {
			log.Fatalln(encodeErr)
		}
		return
	}
}
