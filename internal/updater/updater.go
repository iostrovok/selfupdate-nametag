package updater

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/minio/selfupdate"
	"github.com/pkg/errors"

	"nametag/internal/imagestore"
	"nametag/internal/lg"
)

const (
	// CheckURL is the URL to check for updates
	// todo: move to configuration
	CheckURL = "http://127.0.0.1:8080"

	// ScanFrequency specifies how often request new image. For test it's set up to 10 second.
	// todo: move to configuration
	ScanFrequency = 10 * time.Second
)

var (
	NetError          = errors.Errorf("net error")
	CheckVersionError = errors.Errorf("check versition error")
	RunError          = errors.Errorf("run app error")
)

// Verifier checks binaryData signature by public key
type Verifier interface {
	Verify(binaryData, signature string) error
}

type Updater struct {
	// current process parameters for passing to the new process
	pwdDir          string
	execName        string
	commandLineArgs []string

	// objects to check and identify the new version
	verifier       Verifier
	currentVersion *version.Version

	// Logger for logging. No more than one logger is needed.
	log *lg.Logger
}

func New(log *lg.Logger, ver Verifier, currentVersion string) (*Updater, error) {
	// Get the current process exe file and directory.
	pwdDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	execName, err := os.Executable()
	if err != nil {
		return nil, err
	}

	args := make([]string, len(os.Args))
	copy(args, os.Args)

	c, err := version.NewVersion(currentVersion)
	if err != nil {
		return nil, err
	}

	return &Updater{
		log:             log,
		verifier:        ver,
		currentVersion:  c,
		commandLineArgs: args,
		pwdDir:          pwdDir,
		execName:        execName,
	}, nil
}

// errorHandler is an example of an error handler.
func (u *Updater) errorHandler(err error) {
	if err == nil {
		return
	}

	// todo: add error handling
	// todo: info user about error
	// check NetError, CheckVersionError, RunError
	u.log.Printf("check failed with error: %s", err.Error())
}

// Check checks for new versions of the program and updates it.
// It returns true if new process is success run
// It's a blocking function.
func (u *Updater) Check(ctx context.Context) bool {
	success, err := u.checkAndRun()
	u.errorHandler(err)

	if success {
		return true
	}

	t := time.NewTicker(ScanFrequency)
	for {
		select {
		case <-ctx.Done():
			return false
		case <-t.C:
			success, err := u.checkAndRun()
			u.errorHandler(err)

			if success {
				return true
			}
		}
	}
}

func (u *Updater) checkAndRun() (bool, error) {
	im, err := u.checkNewVersion()
	if err != nil {
		return false, errors.Wrapf(CheckVersionError, err.Error())
	}

	if im == nil {
		return false, nil
	}

	if err := u.loadNewVersion(im); err != nil {
		return false, errors.Wrapf(NetError, err.Error())
	}

	success, err := u.runNext()
	if err != nil {
		err = errors.Wrap(RunError, err.Error())
	}

	return success, err
}

func (u *Updater) checkNewVersion() (*imagestore.Image, error) {
	resp, err := http.Get(CheckURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	im := &imagestore.Image{}
	if err := json.Unmarshal(b, im); err != nil {
		return nil, err
	}

	if im.Version.Compare(u.currentVersion) < 1 {
		return nil, nil
	}

	if err := u.verifier.Verify(im.FileSum, im.Sign); err != nil {
		return nil, err
	}

	return im, nil
}

func (u *Updater) loadNewVersion(im *imagestore.Image) error {
	signB, err := base64.URLEncoding.DecodeString(im.FileSum)
	if err != nil {
		return err
	}

	uri, _ := url.JoinPath(CheckURL, im.Uri)
	resp2, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer resp2.Body.Close()

	// pass the sign of file to check it
	err = selfupdate.Apply(resp2.Body, selfupdate.Options{
		Checksum: signB,
	})

	if err != nil {
		return err
	}

	return nil
}

// runNext starts the next version of the process.
// It does not support any delays or timeouts.
// We assume that all connections, sockets, files, etc. can be used together.
// If you need to add a delay or reuse files, you need to pass and process them here.
// In addition, it may be necessary to update the command arguments (c.Args)
// to delay a new process while the current process closes connections, files, logs, etc.
func (u *Updater) runNext() (bool, error) {
	c := exec.Command(u.execName)
	c.Args = u.commandLineArgs
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	// Start the command.
	if err := c.Start(); err != nil {
		return false, err
	}

	// Wait for the command to finish.
	if c.Process != nil {
		u.log.Infof("start new proccess pid: %d", c.Process.Pid)
	} else {
		return false, errors.New("process is nil")
	}

	return true, nil
}

/*
runNext2 is a low level version of runNext.
I didn't test it because I have no windows in my home.
func (u *Updater) runNext2() (bool, error) {
	// Pass stdin, stdout, and stderr to the child.
	files := []*os.File{
		os.Stdin,
		os.Stdout,
		os.Stderr,
		nil,
	}

	// Spawn child process.
	p, err := os.StartProcess(u.execName, u.commandLineArgs, &os.ProcAttr{
		Dir:   u.pwdDir,
		Env:   os.Environ(),
		Files: files,
		Sys:   &syscall.SysProcAttr{},
	})
	if err != nil {
		return false, err
	}

	u.log.Infof("start new proccess pid: %d", p.Pid)
	if err := p.Release(); err != nil {
		return false, err
	}

	return true, nil
}
*/
