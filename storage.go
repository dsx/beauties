package beauties

import (
	"crypto/sha512"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

// Storage is a Storage interface
type Storage interface {
	String() string
	Get(token, filename string) (reader File, contentType string, contentLength int64, err error)
	Head(token, filename string) (contentType string, contentLength int64, err error)
	Put(token, filename string, reader io.Reader, contentLength int64) error
	Delete(token, filename string) (err error)
	IsNotExist(err error) bool
}

// File is a File interface combining io.Reader, io.Seeker and io.Closer
type File interface {
	io.Reader
	io.Seeker
	io.Closer
}

// LocalStorage is an implementation of a Storage interface using
// local directory to store files
type LocalStorage struct {
	Storage
	Name    string
	basedir string
}

// NewLocalStorage returns a LocalStorage instance
func NewLocalStorage(basedir string) (storage *LocalStorage, err error) {
	storage = &LocalStorage{basedir: basedir, Name: fmt.Sprintf("LocalStorage %s", basedir)}
	err = os.MkdirAll(basedir, 0750)
	return
}

func (s *LocalStorage) hash(filename string) (hash string) {
	fn := []byte(filename)
	hasher := sha512.New()
	hasher.Write(fn)
	hash = fmt.Sprintf("%x", hasher.Sum(nil))
	return
}

func (s *LocalStorage) getPath(token, filename string) (path string) {
	path = filepath.Join(s.basedir, s.hash(token+filename))
	return
}

// String returns a string representation of LocalStorage
func (s *LocalStorage) String() string {
	return s.Name
}

// Head returns content type and content length to use in e.g. HTTP
// HEAD method
func (s *LocalStorage) Head(token string, filename string) (contentType string, contentLength int64, err error) {
	path := s.getPath(token, filename)

	var fi os.FileInfo
	if fi, err = os.Lstat(path); err != nil {
		return
	}

	contentLength = int64(fi.Size())
	contentType = s.getContentType(token, filename)

	return
}

// Get retrieves file from a storage
func (s *LocalStorage) Get(token string, filename string) (reader File, contentType string, contentLength int64, err error) {
	path := s.getPath(token, filename)

	// content type , content length
	if reader, err = os.Open(path); err != nil {
		return
	}

	var fi os.FileInfo
	if fi, err = os.Lstat(path); err != nil {
		return
	}

	contentLength = int64(fi.Size())
	contentType = s.getContentType(token, filename)

	return
}

// Delete deletes file from a storage
func (s *LocalStorage) Delete(token, filename string) (err error) {
	path := s.getPath(token, filename)
	err = os.RemoveAll(path)
	return
}

// IsNotExist checks whether error meaning is file doesn't exists
func (s *LocalStorage) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

// Put puts file in a storage
func (s *LocalStorage) Put(token string, filename string, reader io.Reader, contentLength int64) error {
	var f io.WriteCloser
	var err error

	path := s.getPath(token, filename)

	if f, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600); err != nil && !os.IsExist(err) {
		fmt.Printf("%s", err)
		return err
	}

	defer f.Close()

	if _, err = io.Copy(f, reader); err != nil {
		return err
	}

	return nil
}

func (s *LocalStorage) getContentType(token, filename string) (ct string) {
	if filepath.Ext(filename) == "" {
		return
	}

	ct = mime.TypeByExtension(filepath.Ext(filename))
	if ct != "" {
		return
	}

	var reader File
	var err error
	path := s.getPath(token, filename)

	reader, err = os.Open(path)
	if reader, err = os.Open(path); err != nil {
		return
	}

	buffer := make([]byte, 512)
	if _, err = reader.Read(buffer); err != nil {
		return
	}

	reader.Close()

	ct = http.DetectContentType(buffer)

	return
}
