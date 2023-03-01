package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	// image.Decode won't be able to decode images unless the codecs have been
	// registered by using these side-effecting imports
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/parrotmac/goutil"
)

type distortJob struct {
	Quality       int  `json:"quality"`
	Iterations    int  `json:"iterations"`
	InterleaveGIF bool `json:"interleave_gif"`
}

type req struct {
	Params distortJob `json:"params"`
	Image  string     `json:"image"`
}

func JpegDistort(orig image.Image, out io.Writer, target int, iterations int) error {
	stepSize := int(float64(100-target) / float64(iterations))
	stepSize = goutil.Bounded(stepSize, 1, 10)
	img := orig
	buf := &bytes.Buffer{}
	for i := 100; i >= target; i -= stepSize {
		q := goutil.Bounded(i, 1, 100)
		err := jpeg.Encode(buf, img, &jpeg.Options{
			Quality: q,
		})
		if err != nil {
			return err
		}
		img, err = jpeg.Decode(buf)
		if err != nil {
			return err
		}
		buf.Reset()
	}
	return jpeg.Encode(out, img, &jpeg.Options{})
}

func requestImage(ctx context.Context, imageURL string) (io.Reader, error) {
	httpClient := *(&http.DefaultClient)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/5billion Safari/6trillion")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("encountered err while fetching resource: %s", err.Error())
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("bad status code from upstream server " + resp.Status)
	}

	// 10 MB is a pretty big image
	if resp.ContentLength > 10_000_000 {
		return nil, errors.New("requested image is too large")
	}

	return resp.Body, nil
}

func resolveImage(ctx context.Context, imgData string) (image.Image, error) {

	// note the `== nil`
	if targetURL, err := url.Parse(imgData); err == nil {
		// specified image is really a URL -- go fetch!

		imgData, err := requestImage(ctx, targetURL.String())
		if err != nil {
			return nil, err
		}

		img, _, err := image.Decode(imgData)
		return img, err
	}

	dataParts := strings.Split(imgData, ";base64,")
	if len(dataParts) < 2 {
		return nil, errors.New("invalid data")
	}

	b64decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewBuffer([]byte(dataParts[1])))

	img, _, err := image.Decode(b64decoder)
	return img, err
}

func extractQueryParams(w http.ResponseWriter, r *http.Request) *req {
	var errs []error

	targetURLParam := r.URL.Query().Get("image_url")
	if _, err := url.Parse(targetURLParam); err != nil {
		errs = append(errs, err)
	}

	iterationsParam := r.URL.Query().Get("iterations")
	if iterationsParam == "" {
		iterationsParam = "1"
	}
	iterations, err := strconv.ParseInt(iterationsParam, 10, 64)
	if err != nil {
		errs = append(errs, err)
	}
	if goutil.Outside(iterations, 1, 10) {
		errs = append(errs, errors.New("iterations must be between 1 and 10 (inclusive)"))
	}

	qualityParam := r.URL.Query().Get("quality")
	if qualityParam == "" {
		qualityParam = "50"
	}
	quality, err := strconv.ParseInt(qualityParam, 10, 64)
	if err != nil {
		errs = append(errs, err)
	}
	if goutil.Outside(quality, 1, 100) {
		errs = append(errs, errors.New("quality must be between 1 and 100 (inclusive)"))
	}

	interleaveGif := strings.ToLower(r.URL.Query().Get("interleave_gif")) == "true"

	if len(errs) > 0 {
		errStr := "encountered one or more errors: "
		for _, e := range errs {
			errStr += e.Error()
		}
		http.Error(w, errStr, http.StatusBadRequest)
	}

	return &req{
		Params: distortJob{
			Quality:       int(quality),
			Iterations:    int(iterations),
			InterleaveGIF: interleaveGif,
		},
		Image: targetURLParam,
	}
}

func handleConversion(w http.ResponseWriter, r *http.Request) {
	request := &req{}

	switch r.Method {
	case http.MethodGet:
		request = extractQueryParams(w, r)
		if request == nil {
			return
		}
	case http.MethodPost:
		if err := json.NewDecoder(r.Body).Decode(request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "must send GET or POST request", http.StatusMethodNotAllowed)
		return
	}

	reqCtx, cancelCtx := context.WithTimeout(r.Context(), time.Second*10)
	defer cancelCtx()

	sourceImage, err := resolveImage(reqCtx, request.Image)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = distort(sourceImage, request.Params, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func distort(img image.Image, job distortJob, out io.Writer) error {
	var outgoingImage image.Image
	for i := 0; i < job.Iterations; i++ {
		imgBuffer := &bytes.Buffer{}

		var err error
		if job.InterleaveGIF && i%2 == 0 {
			err = gif.Encode(imgBuffer, img, &gif.Options{})
		} else {
			err = jpeg.Encode(imgBuffer, img, &jpeg.Options{Quality: job.Quality})
		}
		if err != nil {
			return err
		}

		outgoingImage, _, err = image.Decode(imgBuffer)
		if err != nil {
			return err
		}
	}

	enc := base64.NewEncoder(base64.StdEncoding, out)
	return jpeg.Encode(enc, outgoingImage, &jpeg.Options{
		Quality: 100,
	})
}

func getBindAddress() string {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "5000"
	}

	return fmt.Sprintf("0.0.0.0:%s", port)
}

func httpServer() {
	server := http.NewServeMux()

	server.HandleFunc("/api/convert", handleConversion)
	server.Handle("/", http.FileServer(http.Dir("static")))

	bindAddr := getBindAddress()
	log.Printf("Starting HTTP Server at http://%s\n", bindAddr)
	log.Fatalln(http.ListenAndServe(bindAddr, server))
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "server" {
		httpServer()
		return
	}

	cli()
}
