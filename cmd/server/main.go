package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"nametag/internal/imagestore"
	"nametag/internal/signature/sign"
)

const (
	// StorageDir specifies the directory where the images are stored.
	// todo: move to configuration
	StorageDir = "./data"

	// HttpDir specifies the uri to file storage in server
	// todo: move to configuration
	HttpDir = "/data"

	// ScanFrequency specifies how often the image repository should check the catalog for new images.
	// todo: move to configuration
	ScanFrequency = 2 * time.Second
)

type countHandler struct {
	im *imagestore.AllImages
}

func (h *countHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, HttpDir) {
		http.ServeFile(w, r, r.URL.Path[1:])
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err := fmt.Fprintf(w, "%s\n\n", h.im.LastImage)
	if err != nil {
		log.Println(err)
	}
}

func main() {
	singer, err := sign.New()
	if err != nil {
		log.Fatal(err)
	}

	im := imagestore.New(HttpDir, StorageDir, singer)
	ctx, cancel := context.WithCancel(context.Background())
	im.SetScanFrequency(ScanFrequency)

	srv := &http.Server{}
	srv.Handler = &countHandler{im: im}
	srv.Addr = ":8080"

	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		err := im.ScanImages(egCtx)
		if err != nil {
			_ = srv.Shutdown(egCtx)
		}

		return err
	})

	eg.Go(func() error {
		defer cancel()
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	log.Println(eg.Wait())
}
