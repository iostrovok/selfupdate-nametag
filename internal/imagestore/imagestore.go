package imagestore

// AllImages is a struct that holds all the images in the image repository.
// It provides methods to add new images and scan the image directory for new images.
// For each added file, it calculates a sha256 hash and signs it with a private key.

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
)

const (
	// DefaultScanFrequency specifies how often the image repository should check the catalog for new images.
	// todo: move to configuration
	DefaultScanFrequency = 10 * time.Second
)

var releaseVersion = regexp.MustCompile(`\.(v[0-9.]+)$`)

type Signer interface {
	SignFile(string) ([]byte, []byte, error)
}

type Image struct {
	Uri       string           `json:"uri"`
	Image     string           `json:"image"`
	CreatedAt string           `json:"created_at"`
	FileSum   string           `json:"file_sum"`
	Sign      string           `json:"sign"`
	Version   *version.Version `json:"version"`
}

type AllImages struct {
	mx sync.RWMutex

	httpDir       string
	dir           string
	Sing          Signer
	Images        map[string]Image
	LastImage     string // json string for quick access to last image
	scanFrequency time.Duration
}

func New(httpDir, dir string, sign Signer) *AllImages {
	return &AllImages{
		httpDir:       httpDir,
		dir:           dir,
		Sing:          sign,
		Images:        map[string]Image{},
		scanFrequency: DefaultScanFrequency,
	}
}

func (im *AllImages) SetScanFrequency(scanFrequency time.Duration) {
	im.scanFrequency = scanFrequency
}

func (im *AllImages) CheckFile(fileName string) bool {
	im.mx.RLock()
	defer im.mx.RUnlock()

	_, find := im.Images[fileName]
	return find
}

func (im *AllImages) AddFile(fileName string) error {
	ver, err := GetVersion(fileName)
	if err != nil {
		return err
	}

	fullName := path.Join(im.dir, fileName)

	f, err := os.Open(fullName)
	if err != nil {
		return err
	}

	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	defer f.Close()

	sign, fileSign, err := im.Sing.SignFile(fullName)
	if err != nil {
		return err
	}

	im.mx.Lock()
	defer im.mx.Unlock()

	im.Images[fileName] = Image{
		Uri:       path.Join(im.dir, fileName),
		Image:     fileName,
		FileSum:   base64.URLEncoding.EncodeToString(sign),
		Sign:      base64.URLEncoding.EncodeToString(fileSign),
		CreatedAt: time.Now().Format(time.DateTime),
		Version:   ver,
	}

	oldLastImage := im.LastImage
	lastImage := im.Images[fileName]
	for _, f := range im.Images {
		if lastImage.Version.LessThan(f.Version) {
			lastImage = f
		}
	}

	b, err := json.Marshal(lastImage)
	if err != nil {
		return err
	}

	im.LastImage = string(b)
	if oldLastImage != im.LastImage {
		log.Printf("Added new file: %s", im.LastImage)
	}

	return nil
}

// GetVersion extracts the version from the file name
// it's a simple helper.
func GetVersion(fileName string) (*version.Version, error) {
	_, file := filepath.Split(fileName)

	s := releaseVersion.FindStringSubmatch(file)
	if len(s) != 2 {
		return nil, errors.Errorf("not found version in %s", file)
	}

	return version.NewVersion(s[1])
}

// ScanImages scans the image directory for new images
// at a defined frequency until the provided context is canceled.
// It's a wrapper over ScanImagesInDir
func (im *AllImages) ScanImages(ctx context.Context) error {
	// run it at startup to get early errors
	if err := im.ScanImagesInDir(); err != nil {
		return err
	}

	t := time.NewTicker(im.scanFrequency)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			if err := im.ScanImagesInDir(); err != nil {
				return err
			}
		}
	}
}

// ScanImagesInDir checks the specified directory for executable image files
// and adds new ones to the Images map.
func (im *AllImages) ScanImagesInDir() error {
	entries, err := os.ReadDir(im.dir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			return err
		}

		perm := info.Mode().Perm()
		executed := perm&0100 == 1 || perm&0010 == 1 || perm&0001 == 1

		if e.IsDir() || !executed {
			continue
		}

		if im.CheckFile(e.Name()) {
			continue
		}

		if err := im.AddFile(e.Name()); err != nil {
			return err
		}
	}

	return nil
}
