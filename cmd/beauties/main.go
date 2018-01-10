//go:generate go-bindata static

package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/dsx/beauties"
	"github.com/gorilla/mux"
)

// BinaryName defines name of the binary for help message, default
// storage dir name, prefix for temporary files, prefix for some log
// messages and environment variable (converted to uppercase)
var BinaryName = "beauties"

// RequestMaximumMemory maximum size of request to be processed in memory
var RequestMaximumMemory int64 = 1 * 1024 * 1024

var config struct {
	Bind       string
	DomainName string
	Log        string
	MinSpace   float64
	Storage    string
	Temp       string
}

var storage beauties.Storage

func init() {
	// Bind
	config.Bind = ":8080"
	flag.StringVar(&config.Bind, "b", config.Bind, "IP:PORT to bind to")

	// Log
	flag.StringVar(&config.Log, "l", "", "log file (default \"stderr\")")

	// MinSpace
	flag.Float64Var(&config.MinSpace, "m", 0.2, "minimum free space 0…1")

	// Storage
	config.Storage = os.Getenv(fmt.Sprintf("%s_STORAGE", strings.ToUpper(BinaryName)))
	if config.Storage == "" {
		config.Storage = fmt.Sprintf("/var/cache/%s", BinaryName)
	}
	flag.StringVar(&config.Storage, "s", config.Storage, "storage directory")

	// Temp
	flag.StringVar(&config.Temp, "t", filepath.Join(config.Storage, "tmp"), "temporary directory")

	// DomainName
	config.DomainName = os.Getenv(fmt.Sprintf("%s_DOMAIN", strings.ToUpper(BinaryName)))
	flag.StringVar(&config.DomainName, "d", config.DomainName, "domain name to use in response (default is guess from request)")

	flag.Usage = usage
	flag.Parse()
}

func usage() {
	fmt.Printf("Usage: %s [-b 192.0.2.1:80] [-d example.com] [-l /var/log/%s] [-m 0.2] [-s /var/cache/%s] [-t /tmp]\n", BinaryName, BinaryName, BinaryName)
	flag.PrintDefaults()
	fmt.Println("")
}

func main() {
	var err error
	rand.Seed(time.Now().UTC().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())

	r := mux.NewRouter()
	registerHandlers(r)

	if config.Log != "" {
		f, err := os.OpenFile(config.Log, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("Can't open log file %s for writing: %s", config.Log, err.Error())
		}

		defer f.Close()

		log.SetOutput(f)
	}

	if config.Temp != "" {
		os.Setenv("TMPDIR", config.Temp)
	} else {
		os.Setenv("TMPDIR", os.TempDir())
	}

	storage, err = beauties.NewLocalStorage(config.Storage)
	if err != nil {
		log.Fatalf("Can't initialize storage %s: %s", config.Storage, err.Error())
	}

	mime.AddExtensionType(".md", "text/x-markdown")

	log.Printf("%s server started: listening on %s, using temp directory %s, storing files to %s",
		BinaryName, config.Bind, config.Temp, config.Storage)

	s := &http.Server{
		Addr:         config.Bind,
		Handler:      logRequest(r),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	go func() {
		s.ListenAndServe()
	}()

	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt)
	signal.Notify(term, syscall.SIGTERM)

	<-term

	log.Printf("%s server stopped.", BinaryName)
}
