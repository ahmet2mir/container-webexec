package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"strings"

	log "github.com/sirupsen/logrus"
)

func parseTimeout(timeout string) time.Duration {
	if t, err := time.ParseDuration(timeout); err == nil {
		return t
	}
	return time.Duration(0)
}

func execCommand(script string, args string, timeout time.Duration) (string, error) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	logger.WithFields(log.Fields{"script": script}).Info("execCommand(): Script")
	logger.WithFields(log.Fields{"args": args}).Info("execCommand(): Args")
	logger.WithFields(log.Fields{"timeout": timeout}).Info("execCommand(): Timeout Value")

	if timeout != time.Duration(0) {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	a := strings.Split(args, " ")

	if v, err := exec.CommandContext(ctx, script, a...).CombinedOutput(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logger.WithFields(log.Fields{"error": err.Error()}).Error("execCommand(): Timeout Exceed")
			return "", fmt.Errorf("execCommand(): TimeoutError '%v', output '%v", err, string(v))
		} else {
			logger.WithFields(log.Fields{"error": err.Error()}).Error("execCommand(): Non-zero exit code")
			return "", fmt.Errorf("execCommand(): ExitCodeNotZero '%v', output '%v", err, string(v))
		}
	} else {
		return string(v), nil
	}
}

func saveFile(file io.Reader, dir string, name string, mode os.FileMode) (string, error) {
	fileDest := filepath.Join(dir, name)
	logger.WithFields(log.Fields{"fileDest": fileDest}).Info("saveFile(): Destination")

	// Read file content
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("saveFile(): ReadAllError '%v'", err)
	}

	// Create Base Directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("saveFile(): MkdirAll '%v'", err)
	}

	// Write content
	if err := ioutil.WriteFile(fileDest, fileBytes, mode); err != nil {
		return "", fmt.Errorf("saveFile(): WriteFileError '%v'", err)
	}

	return fileDest, nil
}
