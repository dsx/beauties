package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"syscall"

	"github.com/dsx/beauties"
)

func makeURL(r *http.Request, token, filename string) string {
	scheme := "http"
	domain := getDomain(r)
	if r.TLS != nil || r.Header.Get("X-HTTPS") != "" {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s/%s/%s", scheme, domain, token, filename)
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
