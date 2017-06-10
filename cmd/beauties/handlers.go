package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

func registerHandlers(r *mux.Router) {
	r.HandleFunc("/", postHandler).Methods("POST")
	r.HandleFunc("/", viewHandler).Methods("GET")
	r.HandleFunc("/bash", bashHandler).Methods("GET")
	r.HandleFunc("/f", formHandler).Methods("GET")
	r.HandleFunc("/gpg.asc", gpgHandler).Methods("GET")
	r.HandleFunc("/ip", ipHandler).Methods("GET")
	r.HandleFunc("/{filename}", putHandler).Methods("PUT")
	r.HandleFunc("/{token}/{filename}", deleteHandler).Methods("DELETE")
	r.HandleFunc("/{token}/{filename}", getHandler).Methods("GET")
	r.HandleFunc("/{token}/{filename}", headHandler).Methods("HEAD")
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func staticAssetHandler(w http.ResponseWriter, asset string) {
	data, err := Asset(asset)
	if err != nil {
		log.Printf("Can't retrieve asset %s: %s", asset, err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(data); err != nil {
		log.Printf("Can't write asset %s data: %s", asset, err.Error())
		w.Header().Set("Content-Type", "text/html")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	asset := "static/index.txt"
	domain := getDomain(r)
	data, err := Asset(asset)
	if err != nil {
		log.Printf("Can't retrieve asset %s: %s", asset, err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	tmpl, err := template.New("index.txt").Parse(string(data))
	w.Header().Set("Content-Type", "text/plain")
	tmpl.Execute(w, map[string]string{"domain": domain})
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	staticAssetHandler(w, "static/form.html")
}

func ipHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	remoteAddr := r.Header.Get("X-Real-IP")
	if remoteAddr == "" {
		remoteAddr = r.RemoteAddr
		idx := strings.LastIndex(remoteAddr, ":")
		if idx != -1 {
			remoteAddr = remoteAddr[:idx]
		}
	}
	fmt.Fprintf(w, "%s\n", remoteAddr)
}

func gpgHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	staticAssetHandler(w, "static/gpg.asc")
}

func bashHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	staticAssetHandler(w, "static/bash")
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if !checkFreeSpace() {
		log.Printf("%s: not enough free space to store file", storage)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := r.ParseMultipartForm(RequestMaximumMemory); nil != err {
		log.Printf("Can't parse form: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	for _, fheaders := range r.MultipartForm.File {
		for _, fheader := range fheaders {
			filename := filepath.Base(strings.TrimSpace(fheader.Filename))
			contentType := fheader.Header.Get("Content-Type")

			if contentType == "" {
				contentType = mime.TypeByExtension(filepath.Ext(filename))
			}

			fh, err := fheader.Open()
			if err != nil {
				log.Printf("Can't get filehandler for uploaded file %s: %s", filename, err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			token, contentLength, err := genericUploadHandler(fh, filename, contentType)
			if err != nil {
				log.Printf("File upload failed for file %s: %s", filename, err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			url, err := makeURL(r, token, filename)
			if err != nil {
				log.Printf("Can't make url for uploaded file %s: %s", filename, err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			log.Printf("Upload complete: %s (%d)", url, contentLength)
			io.WriteString(w, url+"\n")
		}
	}
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	if !checkFreeSpace() {
		log.Printf("%s: not enough free space to store file", storage)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	filename := filepath.Base(strings.TrimSpace(vars["filename"]))

	contentType := r.Header.Get("Content-Type")

	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(vars["filename"]))
	}

	fh := r.Body

	token, contentLength, err := genericUploadHandler(fh, filename, contentType)
	if err != nil {
		log.Printf("File upload failed for file %s: %s", filename, err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	url, err := makeURL(r, token, filename)
	if err != nil {
		log.Printf("Can't make url for uploaded file %s: %s", filename, err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Printf("Upload complete: %s (%d)", url, contentLength)
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, url+"\n")
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	fileManipulationHandler("Get", w, r)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	fileManipulationHandler("Delete", w, r)
}

func headHandler(w http.ResponseWriter, r *http.Request) {
	fileManipulationHandler("Head", w, r)
}

func genericUploadHandler(fh io.Reader, fn, contentType string) (token string, contentLength int64, err error) {
	filename := filepath.Base(fn)
	token = getToken()

	if contentType == "" {
		contentType = mime.TypeByExtension(filename)
	}

	var buff bytes.Buffer

	contentLength, err = io.CopyN(&buff, fh, RequestMaximumMemory+1)
	if err != nil && err != io.EOF {
		log.Printf("Can't write to buffer: %s", err.Error())
		return
	}

	var reader io.Reader
	if contentLength > RequestMaximumMemory {
		var file *os.File
		file, err = ioutil.TempFile(config.Temp, fmt.Sprintf("%s-", BinaryName))

		if err != nil {
			log.Printf("Can't create temp file: %s", err.Error())
			return
		}
		defer file.Close()

		contentLength, err = io.Copy(file, io.MultiReader(&buff, fh))
		if err != nil {
			log.Printf("Can't write to temp file %s: %s", file.Name(), err.Error())
			os.Remove(file.Name())
			return
		}

		reader, err = os.Open(file.Name())
		if err != nil {
			log.Printf("Can't open temp file %s for reading: %s", file.Name(), err.Error())
			os.Remove(file.Name())
			return
		}
	} else {
		reader = bytes.NewReader(buff.Bytes())
	}

	if err = storage.Put(token, filename, reader, contentType, contentLength); err != nil {
		log.Printf("Can't write to storage %s: %s", storage, err.Error())
	}

	return
}

func fileManipulationHandler(op string, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]
	filename := vars["filename"]
	var (
		contentLength int64
		contentType   string
		err           error
		reader        io.ReadCloser
	)

	switch op {
	case "Get":
		reader, contentType, contentLength, err = storage.Get(token, filename)
		if err == nil {
			defer reader.Close()
		}

	case "Delete":
		err = storage.Delete(token, filename)

	case "Head":
		contentType, contentLength, err = storage.Head(token, filename)
	}

	if err != nil {
		if storage.IsNotExist(err) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		log.Printf("Can't %s file %s/%s from storage %s: %s", op, token, filename, storage, err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return

	}

	switch op {
	case "Get":
		if _, err = io.Copy(w, reader); err != nil {
			log.Printf("Can't read file %s/%s from storage %s: %s", token, filename, storage, err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		fallthrough

	case "Head":
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
		w.Header().Set("Connection", "close")

	case "Delete":
		http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
	}
}
