package rtsp

import (
	"github.com/CVDS2020/CVDS2020/common/assert"
	"github.com/CVDS2020/CVDS2020/common/log"
	"github.com/CVDS2020/CVDS2020/cvds-mdu/config"
	"net"
	"sync"
)

type Server struct {
	//lifecycle.Lifecycle
	//runner *lifecycle.DefaultRunner

	listener *net.TCPListener
	addr     *net.TCPAddr
	stopped  bool

	pushers     map[string]*Pusher // Path <-> Pusher
	pushersLock sync.RWMutex

	logger *log.Logger
}

//func (s *Server) start() error {
//	listener, err := net.ListenTCP("tcp", s.addr)
//	if err != nil {
//		return s.logger.ErrorWith("rtsp server listen error", err, log.String("addr", s.addr.String()))
//	}
//	s.listener = listener
//	s.logger.Info("rtsp server start", log.String("addr", s.addr.String()))
//	return nil
//}

//func (s *Server) run() error {
//	for true {
//		conn, err := s.listener.AcceptTCP()
//		if err != nil {
//			s.handleAcceptError(err)
//		}
//		if err := conn.SetReadBuffer(config.RtspConfig().ReadBuffer); err != nil {
//			s.logger.ErrorWith("rtsp server conn set read buffer error", err)
//		}
//		if err := conn.SetWriteBuffer(config.RtspConfig().ReadBuffer); err != nil {
//			s.logger.ErrorWith("rtsp server conn set write buffer error", err)
//		}
//		session := NewSession(s, conn)
//		go session.Start()
//	}
//	panic("impossible")
//}

//func (s *Server) close() error {
//	return s.listener.Close()
//}

//func (s *Server) handleAcceptError(e error) (err error) {
//	err = e
//	s.runner.Lock()
//	defer func() {
//		// clean server listener
//		s.listener = nil
//		s.runner.Unlock()
//		if err != nil {
//			s.logger.ErrorWith("rtsp server accept tcp error", err, log.String("addr", s.addr.String()))
//		}
//		s.logger.Info("rtsp server closed", log.String("addr", s.addr.String()))
//	}()
//
//	if s.runner.Closing() {
//		// manual close, ignore error
//		err = nil
//		return
//	} else {
//		// accept connection error
//		s.runner.ToClosing()
//	}
//	return
//}

func (s *Server) Start() error {
	listener, err := net.ListenTCP("tcp", s.addr)
	if err != nil {
		return s.logger.ErrorWith("rtsp server listen error", err, log.String("addr", s.addr.String()))
	}

	s.stopped = false
	s.listener = listener
	s.logger.Info("rtsp server start", log.String("addr", s.addr.String()))
	for !s.stopped {
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			if s.stopped {
				return nil
			}
			return s.logger.ErrorWith("rtsp server accept tcp error", err)
		}
		if err := conn.SetReadBuffer(config.RtspConfig().ReadBuffer); err != nil {
			s.logger.ErrorWith("rtsp server conn set read buffer error", err)
		}
		if err := conn.SetWriteBuffer(config.RtspConfig().ReadBuffer); err != nil {
			s.logger.ErrorWith("rtsp server conn set write buffer error", err)
		}
		session := NewSession(s, conn)
		go session.Start()
	}
	return nil
}

func (s *Server) Stop() {
	s.logger.Info("rtsp server stop", log.String("addr", s.addr.String()))
	s.stopped = true
	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
	}
	s.pushersLock.Lock()
	s.pushers = make(map[string]*Pusher)
	s.pushersLock.Unlock()
}

func (s *Server) Addr() *net.TCPAddr {
	return s.addr
}

func (s *Server) AddPusher(pusher *Pusher) bool {
	s.pushersLock.Lock()
	if _, ok := s.pushers[pusher.Path()]; !ok {
		s.pushers[pusher.Path()] = pusher
		s.pushersLock.Unlock()
		go pusher.Start()
		s.logger.Info("pusher start", log.String("pusher", pusher.String()), log.Int("pusher size", len(s.pushers)))
		return true
	}
	s.pushersLock.Unlock()
	return false
}

func (s *Server) TryAttachToPusher(session *Session) (int, *Pusher) {
	s.pushersLock.Lock()
	if pusher, ok := s.pushers[session.Path]; ok {
		if pusher.RebindSession(session) {
			s.pushersLock.Unlock()
			s.logger.Info("Attached to a pusher")
			return 1, pusher
		} else {
			s.pushersLock.Unlock()
			return -1, nil
		}
	}
	s.pushersLock.Unlock()
	return 0, nil
}

func (s *Server) RemovePusher(pusher *Pusher) {
	s.pushersLock.Lock()
	if _pusher, ok := s.pushers[pusher.Path()]; ok && pusher.ID() == _pusher.ID() {
		delete(s.pushers, pusher.Path())
		s.pushersLock.Unlock()
		s.logger.Info("pusher end", log.String("pusher", pusher.String()), log.Int("pusher size", len(s.pushers)))
		return
	}
	s.pushersLock.Unlock()
}

func (s *Server) GetPusher(path string) (pusher *Pusher) {
	s.pushersLock.RLock()
	pusher = s.pushers[path]
	s.pushersLock.RUnlock()
	return
}

func (s *Server) GetPushers() (pushers map[string]*Pusher) {
	pushers = make(map[string]*Pusher)
	s.pushersLock.RLock()
	for k, v := range s.pushers {
		pushers[k] = v
	}
	s.pushersLock.RUnlock()
	return
}

func (s *Server) GetPusherSize() (size int) {
	s.pushersLock.RLock()
	size = len(s.pushers)
	s.pushersLock.RUnlock()
	return
}

var server *Server
var serverInitializer sync.Once

func GetServer() *Server {
	if server != nil {
		return server
	}
	serverInitializer.Do(func() {
		server = &Server{
			addr:    config.RtspConfig().GetAddr(),
			stopped: true,
			pushers: make(map[string]*Pusher),
			logger:  assert.Must(config.LogConfig().Build("rtsp.server")),
		}
	})
	return server
}
