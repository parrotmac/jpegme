package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

func constructPlaceholder(c color.Color) image.Image {
	width := 50
	height := 50

	img := image.NewRGBA(image.Rectangle{
		Min: image.Point{},
		Max: image.Point{
			X: width,
			Y: height,
		},
	})

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, c)
		}
	}

	return img
}

func handleConversion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "must send GET request", http.StatusMethodNotAllowed)
		return
	}

	var errs []error

	targetURLParam := r.URL.Query().Get("image_url")
	if targetURLParam == "" {
		img := constructPlaceholder(color.White)
		err := png.Encode(w, img)
		if err != nil {
			log.Println(err)
		}
		return
	}
	if _, err := url.Parse(targetURLParam); err != nil {
		errs = append(errs, err)
	}

	levelParam := r.URL.Query().Get("quality")
	if levelParam == "" {
		levelParam = "50"
	}
	level, err := strconv.ParseInt(levelParam, 10, 64)
	if err != nil {
		errs = append(errs, err)
	}
	if level < 1 || level > 100 {
		errs = append(errs, errors.New("level must be between 1 and 100 (inclusive)"))
	}

	if len(errs) > 0 {
		errStr := "encountered one or more errors: "
		for _, e := range errs {
			errStr += e.Error()
		}
		http.Error(w, errStr, http.StatusBadRequest)
	}

	reqCtx, cancelCtx := context.WithTimeout(r.Context(), time.Second*10)
	defer cancelCtx()

	httpClient := *(&http.DefaultClient)

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, targetURLParam, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/5billion Safari/6trillion")

	resp, err := httpClient.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("encountered err while fetching resource: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		http.Error(w, "bad status code from upstream server "+resp.Status, http.StatusInternalServerError)
		return
	}

	// 10 MB is a pretty big image
	if resp.ContentLength > 10_000_000 {
		http.Error(w, "requested image is too large", http.StatusBadRequest)
		return
	}

	// image.Decode won't be able to decode images the codecs have been registered by
	// importing the package (hence the `_ "image/..."` imports above)
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respBuffer := &bytes.Buffer{}
	err = jpeg.Encode(respBuffer, img, &jpeg.Options{
		Quality: int(level),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "image/jpeg")

	_, err = io.Copy(w, respBuffer)
	if err != nil {
		log.Println("failed to write response", err)
	}
	return
}

func getBindAddress() string {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "5000"
	}

	return fmt.Sprintf("0.0.0.0:%s", port)
}

func main() {
	server := http.NewServeMux()

	server.HandleFunc("/api/convert", handleConversion)
	server.Handle("/", http.FileServer(http.Dir("static")))

	bindAddr := getBindAddress()
	log.Printf("Starting HTTP Server at http://%s\n", bindAddr)
	log.Fatalln(http.ListenAndServe(bindAddr, server))
}
