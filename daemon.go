package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

const DAEMON_ENV = "GO_DAEMON"

func runDaemon() {
	environ := os.Environ()
	environ = append(environ, DAEMON_ENV+"=true")
	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Env = environ
	cmd.Start()
	cmdErrC := make(chan error)
	go func() {
		cmdErrC <- cmd.Wait()
	}()
	select {
	case err := <-cmdErrC:
		log.Fatalf("server started failed, %v", err)
	case <-time.After(200 * time.Millisecond):
		fmt.Printf("server started.\n")
	}
}

func setupLogfile(file string) *os.File {
	logfd, err := os.Create(fLogfile)
	if err != nil {
		panic(err)
	}
	defer logfd.Close()
	os.Stdout = logfd
	os.Stderr = logfd
	os.Stdin = nil
	return logfd
}
