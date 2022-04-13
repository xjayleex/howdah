package infra

import (
	"errors"
	fqdn "github.com/Showmax/go-fqdn"
	"os"
	"strings"
)

var cachedHostname string
var cachedPublicHostname string

func Hostname() (string, error) {
	if strings.Compare(cachedHostname, "") != 0 {
		return cachedHostname, nil
	}

	if f, err := fqdn.FqdnHostname(); err == nil {
		cachedHostname = f
		return f, nil
	}

	if hostname, err := os.Hostname(); err == nil {
		cachedHostname = hostname
		return hostname, nil
	}

	return "", errors.New("hostname can not be resolved")
}

func PublicHostname() (string, error) {
	if strings.Compare(cachedPublicHostname, "") != 0 {
		return cachedPublicHostname, nil
	}
	// TODO : What is Public Hostname ? Above all, it returns Hostname()
	return Hostname()
}