package rtsp

import (
	"bytes"
	"fmt"
	"github.com/CVDS2020/CVDS2020/common/log"
	"github.com/CVDS2020/CVDS2020/cvds-mdu/config"
	"net"
	"strconv"
	"strings"
	"time"
)

type UDPServer struct {
	*Session
	*Client

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

func (s *UDPServer) AddInputBytes(bytes int) {
	if s.Session != nil {
		s.Session.InBytes += bytes
		return
	}
	if s.Client != nil {
		s.Client.InBytes += bytes
		return
	}
	panic(fmt.Errorf("session and Client both nil"))
}

func (s *UDPServer) HandleRTP(pack *RTPPack) {
	if s.Session != nil {
		for _, v := range s.Session.RTPHandles {
			v(pack)
		}
		return
	}

	if s.Client != nil {
		for _, v := range s.Client.RTPHandles {
			v(pack)
		}
		return
	}
	panic(fmt.Errorf("session and Client both nil"))
}

func (s *UDPServer) Logger() *log.Logger {
	if s.Session != nil {
		return s.Session.logger
	}
	if s.Client != nil {
		return s.Client.logger
	}
	panic(fmt.Errorf("session and Client both nil"))
}

func (s *UDPServer) Stop() {
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

func (s *UDPServer) SetupAudio() (err error) {
	var (
		logger = s.Logger()
		addr   *net.UDPAddr
	)
	if addr, err = net.ResolveUDPAddr("udp", ":0"); err != nil {
		return
	}
	if s.AConn, err = net.ListenUDP("udp", addr); err != nil {
		return
	}
	if err = s.AConn.SetReadBuffer(config.RtspConfig().Audio.Server.ReadBuffer); err != nil {
		logger.ErrorWith("udp server audio conn set read buffer error", err)
	}
	if err = s.AConn.SetWriteBuffer(config.RtspConfig().Audio.Server.WriteBuffer); err != nil {
		logger.ErrorWith("udp server audio conn set write buffer error", err)
	}
	la := s.AConn.LocalAddr().String()
	strPort := la[strings.LastIndex(la, ":")+1:]
	s.APort, err = strconv.Atoi(strPort)
	if err != nil {
		return
	}
	go func() {
		bufUDP := make([]byte, UdpBufSize)
		logger.Info("udp server start listen audio", log.Int("APort", s.APort))
		defer logger.Info("udp server stop listen audio", log.Int("APort", s.APort))
		timer := time.Unix(0, 0)
		for !s.Stoped {
			if n, _, err := s.AConn.ReadFromUDP(bufUDP); err == nil {
				elapsed := time.Now().Sub(timer)
				if elapsed >= 30*time.Second {
					logger.Debug("Package recv from AConn.len", log.Int("len", n))
					timer = time.Now()
				}
				rtpBytes := make([]byte, n)
				s.AddInputBytes(n)
				copy(rtpBytes, bufUDP)
				pack := &RTPPack{
					Type:   RtpTypeAudio,
					Buffer: bytes.NewBuffer(rtpBytes),
				}
				s.HandleRTP(pack)
			} else {
				logger.ErrorWith("udp server read audio pack error", err)
				continue
			}
		}
	}()
	addr, err = net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return
	}
	s.AControlConn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return
	}
	if err = s.AControlConn.SetReadBuffer(config.RtspConfig().AudioControl.Server.ReadBuffer); err != nil {
		logger.ErrorWith("udp server audio control conn set read buffer error", err)
	}
	if err = s.AControlConn.SetWriteBuffer(config.RtspConfig().AudioControl.Server.WriteBuffer); err != nil {
		logger.ErrorWith("udp server audio control conn set write buffer error", err)
	}
	la = s.AControlConn.LocalAddr().String()
	strPort = la[strings.LastIndex(la, ":")+1:]
	s.AControlPort, err = strconv.Atoi(strPort)
	if err != nil {
		return
	}
	go func() {
		bufUDP := make([]byte, UdpBufSize)
		logger.Info("udp server start listen audio control", log.Int("AControlPort", s.AControlPort))
		defer logger.Info("udp server stop listen audio control", log.Int("AControlPort", s.AControlPort))
		for !s.Stoped {
			if n, _, err := s.AControlConn.ReadFromUDP(bufUDP); err == nil {
				//logger.Printf("Package recv from AControlConn.len:%d\n", n)
				rtpBytes := make([]byte, n)
				s.AddInputBytes(n)
				copy(rtpBytes, bufUDP)
				pack := &RTPPack{
					Type:   RtpTypeAudioControl,
					Buffer: bytes.NewBuffer(rtpBytes),
				}
				s.HandleRTP(pack)
			} else {
				logger.ErrorWith("udp server read audio control pack error", err)
				continue
			}
		}
	}()
	return
}

func (s *UDPServer) SetupVideo() (err error) {
	var (
		logger = s.Logger()
		addr   *net.UDPAddr
	)
	addr, err = net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return
	}
	s.VConn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return
	}
	if err = s.VConn.SetReadBuffer(config.RtspConfig().Video.Server.ReadBuffer); err != nil {
		logger.ErrorWith("udp server video conn set read buffer error", err)
	}
	if err = s.VConn.SetWriteBuffer(config.RtspConfig().Video.Server.WriteBuffer); err != nil {
		logger.ErrorWith("udp server video conn set write buffer error", err)
	}
	la := s.VConn.LocalAddr().String()
	strPort := la[strings.LastIndex(la, ":")+1:]
	s.VPort, err = strconv.Atoi(strPort)
	if err != nil {
		return
	}
	go func() {
		bufUDP := make([]byte, UdpBufSize)
		logger.Info("udp server start listen video", log.Int("VPort", s.VPort))
		defer logger.Info("udp server stop listen video", log.Int("VPort", s.VPort))
		timer := time.Unix(0, 0)
		for !s.Stoped {
			var n int
			if n, _, err = s.VConn.ReadFromUDP(bufUDP); err == nil {
				elapsed := time.Now().Sub(timer)
				if elapsed >= 30*time.Second {
					logger.Debug("Package recv from VConn.len", log.Int("len", n))
					timer = time.Now()
				}
				rtpBytes := make([]byte, n)
				s.AddInputBytes(n)
				copy(rtpBytes, bufUDP)
				pack := &RTPPack{
					Type:   RtpTypeVideo,
					Buffer: bytes.NewBuffer(rtpBytes),
				}
				s.HandleRTP(pack)
			} else {
				logger.ErrorWith("udp server read video pack error", err)
				continue
			}
		}
	}()

	addr, err = net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return
	}
	s.VControlConn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return
	}
	if err = s.VControlConn.SetReadBuffer(config.RtspConfig().VideoControl.Server.ReadBuffer); err != nil {
		logger.ErrorWith("udp server video control conn set read buffer error", err)
	}
	if err = s.VControlConn.SetWriteBuffer(config.RtspConfig().VideoControl.Server.WriteBuffer); err != nil {
		logger.ErrorWith("udp server video control conn set write buffer error", err)
	}
	la = s.VControlConn.LocalAddr().String()
	strPort = la[strings.LastIndex(la, ":")+1:]
	s.VControlPort, err = strconv.Atoi(strPort)
	if err != nil {
		return
	}
	go func() {
		bufUDP := make([]byte, UdpBufSize)
		logger.Info("udp server start listen video control", log.Int("VControlPort", s.VControlPort))
		defer logger.Info("udp server stop listen video control", log.Int("VControlPort", s.VControlPort))
		for !s.Stoped {
			var n int
			if n, _, err = s.VControlConn.ReadFromUDP(bufUDP); err == nil {
				//logger.Printf("Package recv from VControlConn.len:%d\n", n)
				rtpBytes := make([]byte, n)
				s.AddInputBytes(n)
				copy(rtpBytes, bufUDP)
				pack := &RTPPack{
					Type:   RtpTypeVideoControl,
					Buffer: bytes.NewBuffer(rtpBytes),
				}
				s.HandleRTP(pack)
			} else {
				logger.ErrorWith("udp server read video control pack error", err)
				continue
			}
		}
	}()
	return
}
