package config

import (
	"github.com/CVDS2020/CVDS2020/common/config"
	"net"
	"strconv"
	"time"
)

type Http struct {
	// http server listening host, default 127.0.0.1
	Host string `yaml:"host" json:"host"`
	// http server listening port, default 8190
	Port int `yaml:"port" json:"port"`
	// http server listening address, calculate by Host and Port
	addr string

	// http server read timeout, default is http server default
	ReadTimeout time.Duration `yaml:"read-timeout" json:"read-timeout"`
	// http server read header timeout, default is http server default
	ReadHeaderTimeout time.Duration `yaml:"read-header-timeout" json:"read-header-timeout"`
	// http server write timeout, default is http server default
	WriteTimeout time.Duration `yaml:"write-timeout" json:"write_timeout"`
	// http server idle timeout, default is http server default
	IdleTimeout time.Duration `yaml:"idle-timeout" json:"idle-timeout"`

	Gin struct {
		EnableConsoleColor bool `yaml:"enable-console-color" json:"enable-console-color"`
	} `yaml:"gin" json:"gin"`
}

func (h *Http) PreHandle() config.PreHandlerConfig {
	if h == nil {
		h = new(Http)
	}
	// default config value
	h.Host = "127.0.0.1"
	h.Port = 8190
	return h
}

func (h *Http) PostHandle() (config.PostHandlerConfig, error) {
	// calculate http server listening address
	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(h.Host, strconv.Itoa(h.Port)))
	if err != nil {
		return nil, err
	}
	h.addr = addr.String()
	return h, nil
}

func (h *Http) GetAddr() string {
	return h.addr
}
