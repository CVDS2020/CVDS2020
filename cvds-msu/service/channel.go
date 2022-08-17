package service

import (
	"github.com/CVDS2020/CVDS2020/common/assert"
	"github.com/CVDS2020/CVDS2020/common/def"
	"github.com/CVDS2020/CVDS2020/common/errors"
	"github.com/CVDS2020/CVDS2020/common/log"
	"github.com/CVDS2020/CVDS2020/cvds-msu/config"
	"github.com/CVDS2020/CVDS2020/cvds-msu/storage"
	urlpkg "net/url"
	"strings"
	"sync"
)

var (
	InvalidChannelNameError = errors.New("invalid channel name")
	InvalidChannelURLError  = errors.New("invalid channel url")
	ChannelExistError       = errors.New("channel exist")
	ChannelNotFoundError    = errors.New("channel not found")
)

type Channel struct {
	channels     map[string]*storage.Channel
	channelNames map[string]*storage.Channel
	channelsLock sync.Mutex
	logger       *log.Logger
}

func (s *Channel) addChannel(channel *storage.Channel) bool {
	s.channelsLock.Lock()
	defer s.channelsLock.Unlock()
	if _, has := s.channels[channel.UUID()]; has {
		return false
	}
	if _, has := s.channelNames[channel.Name()]; has {
		return false
	}
	s.channels[channel.UUID()] = channel
	s.channelNames[channel.UUID()] = channel
	return true
}

func (s *Channel) removeChannel(id string) *storage.Channel {
	s.channelsLock.Lock()
	defer s.channelsLock.Unlock()
	ch, has := s.channels[id]
	if !has {
		return nil
	}
	delete(s.channels, id)
	delete(s.channels, ch.Name())
	return ch
}

func (s *Channel) CreateChannel(name string, url string, transport string, cover uint, fields map[string]any) (*storage.Channel, error) {
	transport = strings.ToLower(transport)
	switch transport {
	case "tcp", "udp":
	default:
		transport = "tcp"
	}
	up, err := urlpkg.Parse(url)
	if err != nil || up.Scheme != "rtsp" {
		return nil, InvalidChannelURLError
	}
	def.SetDefault(&cover, 24)
	ch := storage.NewChannel(name, url, transport, cover, fields)
	if !s.addChannel(ch) {
		return nil, ChannelExistError
	}
	return ch, nil
}

func (s *Channel) GetChannel(id string) (*storage.Channel, error) {
	s.channelsLock.Lock()
	defer s.channelsLock.Unlock()
	ch, has := s.channels[id]
	if !has {
		return nil, ChannelNotFoundError
	}
	return ch, nil
}

func (s *Channel) ChannelStart(id string) error {
	ch, err := s.GetChannel(id)
	if err != nil {
		return err
	}
	return ch.Start()
}

func (s *Channel) ChannelStop(id string) error {
	ch, err := s.GetChannel(id)
	if err != nil {
		return err
	}
	err, _ = ch.CloseWait()
	return err
}

func (s *Channel) RemoveChannel(id string) error {
	ch := s.removeChannel(id)
	if ch == nil {
		return ChannelNotFoundError
	}
	return ch.Destroy()
}

func (s *Channel) RemoveAll() {
	var removed []*storage.Channel
	s.channelsLock.Lock()
	for id, ch := range s.channels {
		removed = append(removed, ch)
		delete(s.channels, id)
	}
	for name := range s.channelNames {
		delete(s.channelNames, name)
	}
	s.channelsLock.Unlock()
	for _, ch := range removed {
		ch.Destroy()
	}
}

var channel *Channel
var channelInitializer sync.Once

func GetChannel() *Channel {
	if channel != nil {
		return channel
	}
	channelInitializer.Do(func() {
		channel = &Channel{
			channels:     make(map[string]*storage.Channel),
			channelNames: make(map[string]*storage.Channel),
			logger:       assert.Must(config.LogConfig().Build("service.channel")),
		}
	})
	return GetChannel()
}
