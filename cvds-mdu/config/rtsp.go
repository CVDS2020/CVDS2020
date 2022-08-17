package config

import (
	"github.com/CVDS2020/CVDS2020/common/config"
	"github.com/CVDS2020/CVDS2020/common/def"
	"github.com/CVDS2020/CVDS2020/common/unit"
	"net"
	"strconv"
	"time"
)

type ReadWriteBuffer struct {
	ReadBuffer  int `yaml:"read-buffer" json:"read-buffer"`
	WriteBuffer int `yaml:"write-buffer" json:"write-buffer"`
}

type ReaderWriter struct {
	ReaderSize int `yaml:"reader-size" json:"reader-size"`
	WriterSize int `yaml:"writer-size" json:"writer-size"`
}

type AV struct {
	ReadWriteBuffer `yaml:",inline"`
	Client          ReadWriteBuffer `yaml:"client" json:"client"`
	Server          ReadWriteBuffer `yaml:"server" json:"server"`
}

func (a *AV) PostHandle() (config.PostHandlerConfig, error) {
	def.SetDefault(&a.WriteBuffer, a.ReadBuffer)
	def.SetDefault(&a.Client.ReadBuffer, a.ReadBuffer)
	def.SetDefault(&a.Client.WriteBuffer, a.WriteBuffer)
	def.SetDefault(&a.Server.ReadBuffer, a.ReadBuffer)
	def.SetDefault(&a.Server.WriteBuffer, a.WriteBuffer)
	return a, nil
}

type Rtsp struct {
	// rtsp server listening host, default 127.0.0.1
	Host string `yaml:"host" json:"host"`
	// rtsp server listening port, default 8190
	Port int `yaml:"port" json:"port"`
	// rtsp server listening address, calculate by Host and Port
	addr *net.TCPAddr

	ReadWriteBuffer `yaml:",inline"`
	ReaderWriter    `yaml:",inline"`

	Timeout time.Duration `yaml:"timeout" json:"timeout"`

	EnableAuthorization bool `yaml:"enable-authorization" json:"enable-authorization"`
	CloseOld            bool `yaml:"close-old" json:"close-old"`

	Client struct {
		ReaderWriter `yaml:",inline"`
		Timeout      time.Duration
	} `yaml:"client" json:"client"`

	Player struct {
		QueueLimit           uint `yaml:"queue-limit" json:"queue-limit"`
		DropPacketWhenPaused bool `yaml:"drop-packet-when-paused" json:"drop-packet-when-paused"`
	} `yaml:"player" json:"player"`

	Pusher struct {
		DisableGopCache bool `yaml:"disable-gop-cache" json:"disable-gop-cache"`
	} `yaml:"pusher" json:"pusher"`

	Audio        AV `yaml:"audio" json:"audio"`
	AudioControl AV `yaml:"audio-control" json:"audio-control"`
	Video        AV `yaml:"video" json:"video"`
	VideoControl AV `yaml:"video-control" json:"video-control"`

	EnableDebug bool `yaml:"enable-debug" json:"enable-debug"`
}

func (r *Rtsp) PreHandle() config.PreHandlerConfig {
	if r == nil {
		r = new(Rtsp)
	}
	r.Host = "127.0.0.1"
	r.Port = 554
	r.ReadBuffer = unit.MeBiByte
	r.ReaderSize = 200 * unit.KiBiByte
	r.Audio.ReadBuffer = 256 * unit.KiBiByte
	r.Video.ReadBuffer = unit.MeBiByte
	return r
}

func (r *Rtsp) PostHandle() (config.PostHandlerConfig, error) {
	// calculate http server listening address
	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(r.Host, strconv.Itoa(r.Port)))
	if err != nil {
		return nil, err
	}
	r.addr = addr
	def.SetDefault(&r.WriteBuffer, r.WriteBuffer)
	def.SetDefault(&r.WriterSize, r.ReaderSize)
	def.SetDefault(&r.Client.ReaderSize, r.ReaderSize)
	def.SetDefault(&r.Client.WriterSize, r.WriterSize)
	def.SetDefault(&r.Client.Timeout, r.Timeout)

	def.SetDefault(&r.Audio.WriteBuffer, r.Audio.ReadBuffer)
	def.SetDefault(&r.AudioControl.ReadBuffer, r.Audio.ReadBuffer)
	def.SetDefault(&r.AudioControl.WriteBuffer, r.Audio.WriteBuffer)

	def.SetDefault(&r.Video.WriteBuffer, r.Video.ReadBuffer)
	def.SetDefault(&r.VideoControl.ReadBuffer, r.Video.ReadBuffer)
	def.SetDefault(&r.VideoControl.WriteBuffer, r.Video.WriteBuffer)
	return r, nil
}

func (r *Rtsp) GetAddr() *net.TCPAddr {
	return r.addr
}
