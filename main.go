package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
)

var logger *log.Logger

type Config struct {
	address string
	baseDir string
	args    string
	command string
}

// based on https://tutorialedge.net/golang/go-file-upload-tutorial/
// see also https://github.com/mayth/go-simple-upload-server
func (cfg *Config) uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	var (
		err  error
		mode os.FileMode
	)

	r.ParseMultipartForm(100 << 20) // 104857600 - 100MB

	// Get File mode
	if v := r.Form.Get("mode"); v != "" {
		v_int, err := strconv.ParseInt(v, 8, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mode = os.FileMode(v_int)
	} else {
		mode = os.FileMode(0755)
	}

	// FormFile returns the first file for the given key `file`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	logger.Debug(fmt.Sprintf("File Name: %+v\n", handler.Filename))
	logger.Debug(fmt.Sprintf("File Size: %+v\n", handler.Size))
	logger.Debug(fmt.Sprintf("File Mode: %+v\n", mode))
	logger.Debug(fmt.Sprintf("File MIME: %+v\n", handler.Header))

	fileDest, err := saveFile(file, cfg.baseDir, handler.Filename, mode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Successfully Uploaded File to "+fileDest+"\n")

	if v := r.Form.Get("exec"); v == "true" {
		timeout := parseTimeout(r.Form.Get("timeout"))
		args := r.Form.Get("args")

		if v, err := execCommand(fileDest, args, timeout); err == nil {
			fmt.Fprintf(w, v+"\n")
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

func (cfg *Config) execCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	timeout := parseTimeout(r.Form.Get("timeout"))
	args := r.Form.Get("args")
	script := r.Form.Get("script")
	if script == "" {
		http.Error(w, "Url Param 'script' is missing", http.StatusBadRequest)
		return
	}

	if v, err := execCommand(script, args, timeout); err == nil {
		fmt.Fprintf(w, v+"\n")
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func main() {
	logger = log.New()

	baseDir := flag.String("basedir", "/tmp/base-ci", "Directory where files are uploaded.")
	bindHost := flag.String("host", "127.0.0.1", "Bind address.")
	bindPort := flag.Int("port", 8080, "Bind port.")
	command := flag.String("command", "", "Command to run in separate goroutine.")
	args := flag.String("args", "", "Command args, put all args separated with a space inside a single string.")
	logLevelFlag := flag.String("loglevel", "info", "logging level")
	flag.Parse()

	if logLevel, err := log.ParseLevel(*logLevelFlag); err != nil {
		log.WithError(err).Error("failed to parse logging level, so set to default")
	} else {
		logger.Level = logLevel
	}

	address := fmt.Sprintf("%s:%d", *bindHost, *bindPort)

	cfg := Config{
		address: address,
		baseDir: *baseDir,
		args:    *args,
		command: *command,
	}

	http.HandleFunc("/upload", cfg.uploadFile)
	http.HandleFunc("/exec", cfg.execCommand)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	// start http server
	go func() {
		logger.Info(fmt.Sprintf(fmt.Sprintf("Running HTTP server and listen on: %+v\n", address)))
		logger.Fatal(http.ListenAndServe(address, nil))
		wg.Done()
	}()

	// start command
	go func() {
		if *command != "" {
			logger.Info(fmt.Sprintf("Running command '%+v' and args '%+v'\n", *command, *args))

			cmd := exec.Command(*command, *args)
			stdout, _ := cmd.StdoutPipe()
			err := cmd.Start()
			if err != nil {
				logger.Fatal(err)
			}

			buf := make([]byte, 4)
			for {
				n, err := stdout.Read(buf)
				fmt.Print(string(buf[:n]))
				if err == io.EOF {
					break
				}
			}

			cmd.Wait()

			logger.Info("Finsihed command")
		}
		wg.Done()
	}()

	wg.Wait()
}
