package weedcl

import (
	"log"
	"net/url"
	"strconv"
	"strings"
)

type HTTPConfig struct {
	BaseURL *url.URL
}

func NewHTTPCfg(baseURL string) *HTTPConfig {
	u, err := url.Parse(baseURL)
	if err != nil {
		log.Fatalln(err)
	}

	cfg := &HTTPConfig{
		BaseURL: u,
	}
	return cfg
}

func (h *HTTPConfig) Host() string {
	ss := strings.Split(h.BaseURL.Host, ":")
	if len(ss) > 1 {
		return ss[0]
	}
	return h.BaseURL.Host
}

func (h *HTTPConfig) Port() int {
	ss := strings.Split(h.BaseURL.Host, ":")
	var port int
	if len(ss) > 1 {
		port, _ = strconv.Atoi(ss[1])
	}
	port = 80
	return port
}

func (s *HTTPConfig) String() string {
	return s.BaseURL.String()
}
