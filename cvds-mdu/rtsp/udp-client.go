package rtsp

import (
	"fmt"
	"github.com/CVDS2020/CVDS2020/cvds-mdu/config"
	"net"
	"strings"
)

type UDPClient struct {
	*Session

	APort        int
	AConn        *net.UDPConn
	AControlPort int
	AControlConn *net.UDPConn
	VPort        int
	VConn        *net.UDPConn
	VControlPort int
	VControlConn *net.UDPConn

	Stoped bool
}

func (s *UDPClient) Stop() {
	if s.Stoped {
		return
	}
	s.Stoped = true
	if s.AConn != nil {
		s.AConn.Close()
		s.AConn = nil
	}
	if s.AControlConn != nil {
		s.AControlConn.Close()
		s.AControlConn = nil
	}
	if s.VConn != nil {
		s.VConn.Close()
		s.VConn = nil
	}
	if s.VControlConn != nil {
		s.VControlConn.Close()
		s.VControlConn = nil
	}
}

func (c *UDPClient) SetupAudio() (err error) {
	var (
		logger = c.logger
		addr   *net.UDPAddr
	)
	defer func() {
		if err != nil {
			logger.ErrorWith("setup audio error", err)
			c.Stop()
		}
	}()
	host := c.Conn.RemoteAddr().String()
	host = host[:strings.LastIndex(host, ":")]
	if addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, c.APort)); err != nil {
		return
	}
	c.AConn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		return
	}
	if err = c.AConn.SetReadBuffer(config.RtspConfig().Audio.Client.ReadBuffer); err != nil {
		logger.ErrorWith("udp client audio conn set read buffer error", err)
	}
	if err = c.AConn.SetWriteBuffer(config.RtspConfig().Audio.Client.WriteBuffer); err != nil {
		logger.ErrorWith("udp client audio conn set write buffer error", err)
	}

	addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, c.AControlPort))
	if err != nil {
		return
	}
	c.AControlConn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		return
	}
	if err = c.AControlConn.SetReadBuffer(config.RtspConfig().AudioControl.Client.ReadBuffer); err != nil {
		logger.ErrorWith("udp client audio control conn set read buffer error", err)
	}
	if err = c.AControlConn.SetWriteBuffer(config.RtspConfig().AudioControl.Client.WriteBuffer); err != nil {
		logger.ErrorWith("udp client audio control conn set write buffer error", err)
	}
	return
}

func (c *UDPClient) SetupVideo() (err error) {
	var (
		logger = c.logger
		addr   *net.UDPAddr
	)
	defer func() {
		if err != nil {
			logger.ErrorWith("setup video error", err)
			c.Stop()
		}
	}()
	host := c.Conn.RemoteAddr().String()
	host = host[:strings.LastIndex(host, ":")]
	if addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, c.VPort)); err != nil {
		return
	}
	if c.VConn, err = net.DialUDP("udp", nil, addr); err != nil {
		return
	}
	if err = c.VConn.SetReadBuffer(config.RtspConfig().Video.Client.ReadBuffer); err != nil {
		logger.ErrorWith("udp client video conn set read buffer error", err)
	}
	if err = c.VConn.SetWriteBuffer(config.RtspConfig().Video.Client.WriteBuffer); err != nil {
		logger.ErrorWith("udp client video conn set write buffer error", err)
	}

	addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, c.VControlPort))
	if err != nil {
		return
	}
	c.VControlConn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		return
	}
	if err = c.VControlConn.SetReadBuffer(config.RtspConfig().VideoControl.Client.ReadBuffer); err != nil {
		logger.ErrorWith("udp client video control conn set read buffer error", err)
	}
	if err = c.VControlConn.SetWriteBuffer(config.RtspConfig().VideoControl.Client.WriteBuffer); err != nil {
		logger.ErrorWith("udp client video control conn set write buffer error", err)
	}
	return
}

func (c *UDPClient) SendRTP(pack *RTPPack) (err error) {
	if pack == nil {
		err = fmt.Errorf("udp client send rtp got nil pack")
		return
	}
	var conn *net.UDPConn
	switch pack.Type {
	case RtpTypeAudio:
		conn = c.AConn
	case RtpTypeAudioControl:
		conn = c.AControlConn
	case RtpTypeVideo:
		conn = c.VConn
	case RtpTypeVideoControl:
		conn = c.VControlConn
	default:
		err = fmt.Errorf("udp client send rtp got unkown pack type[%v]", pack.Type)
		return
	}
	if conn == nil {
		err = fmt.Errorf("udp client send rtp pack type[%v] failed, conn not found", pack.Type)
		return
	}
	var n int
	if n, err = conn.Write(pack.Buffer.Bytes()); err != nil {
		err = fmt.Errorf("udp client write bytes error, %v", err)
		return
	}
	// logger.Printf("udp client write [%d/%d]", n, pack.Buffer.Len())
	c.Session.OutBytes += n
	return
}
