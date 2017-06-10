package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"syscall"

	"github.com/dsx/beauties"
)

func makeURL(r *http.Request, token, filename string) (string, error) {
	domain := getDomain(r)
	_uri := fmt.Sprintf("http://%s/%s/%s", domain, token, filename)
	u, err := url.Parse(_uri)
	if err != nil {
		return "", err
	}

	if r.TLS != nil || r.Header.Get("X-HTTPS") != "" {
		u.Scheme = "https"
	}

	return u.String(), nil
}

func checkFreeSpace() bool {
	var stat syscall.Statfs_t
	syscall.Statfs(config.Storage, &stat)
	if float64(stat.Bavail)/float64(stat.Blocks) <= config.MinSpace {
		return false
	}
	return true

}

func getToken() string {
	return beauties.Encode(10000000 + int64(rand.Intn(1000000000)))
}

func getDomain(r *http.Request) (domain string) {
	domain = config.DomainName
	if domain == "" {
		domain = r.Host
	}
	return
}
