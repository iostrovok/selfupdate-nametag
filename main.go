package main

// A simple server that returns the version and pid of the application.
// It also checks for updates in the background
// For testing only

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/libp2p/go-reuseport"

	"nametag/internal/lg"
	"nametag/internal/signature/verify"
	"nametag/internal/updater"
)

var (
	Version string

	// LogFile is the path to the log file.
	// We use a separate log file for each version
	LogFile = "./data/logs/%s.log" // todo: move to configuration
	Address = "127.0.0.1:8081"     // todo: move to configuration
)

type simpleHandler struct {
	pid int
}

func (h *simpleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello from PID %d and Version %s\n", h.pid, Version)
}

func prepareServer() (*lg.Logger, *updater.Updater, error) {
	log, err := lg.New(fmt.Sprintf(LogFile, Version), Version)
	if err != nil {
		return nil, nil, err
	}

	log.Infof("start version: %s, pid: %s", Version, os.Getpid())
	ver, err := verify.New()
	if err != nil {
		log.Errorf("error create verify: %s", err.Error())
		return nil, nil, err
	}

	u, err := updater.New(log, ver, Version)
	if err != nil {
		log.Errorf("error create updater: %s", err.Error())
		return nil, nil, err
	}

	return log, u, nil
}

func main() {
	log, u, err := prepareServer()
	if err != nil {
		panic(err)
	}
	defer log.Close()

	l, err := reuseport.Listen("tcp", Address)
	if err != nil {
		log.Fatal(err.Error())
	}

	server := &http.Server{}
	server.Handler = &simpleHandler{pid: os.Getpid()}

	go func() {
		// waiting for the new version
		if u.Check(context.Background()) {
			// shutdown the server
			_ = l.Close()
			_ = server.Close()
		}
	}()

	// run our server
	server.Serve(l)
}
