package rtsp

import (
	"github.com/CVDS2020/CVDS2020/common/log"
	"github.com/CVDS2020/CVDS2020/cvds-mdu/config"
	"sync"
	"time"
)

type Player struct {
	*Session
	Pusher               *Pusher
	cond                 *sync.Cond
	queue                []*RTPPack
	queueLimit           uint
	dropPacketWhenPaused bool
	paused               bool
}

func NewPlayer(session *Session, pusher *Pusher) (player *Player) {
	player = &Player{
		Session:              session,
		Pusher:               pusher,
		cond:                 sync.NewCond(&sync.Mutex{}),
		queue:                make([]*RTPPack, 0),
		queueLimit:           config.RtspConfig().Player.QueueLimit,
		dropPacketWhenPaused: config.RtspConfig().Player.DropPacketWhenPaused,
		paused:               false,
	}
	session.StopHandles = append(session.StopHandles, func() {
		pusher.RemovePlayer(player)
		player.cond.Broadcast()
	})
	return
}

func (player *Player) QueueRTP(pack *RTPPack) *Player {
	logger := player.logger
	if pack == nil {
		logger.Warn("player queue enter nil pack, drop it")
		return player
	}
	if player.paused && player.dropPacketWhenPaused {
		return player
	}
	player.cond.L.Lock()
	player.queue = append(player.queue, pack)
	if oldLen := len(player.queue); player.queueLimit > 0 && oldLen > int(player.queueLimit) {
		player.queue = player.queue[1:]
		if config.RtspConfig().EnableDebug {
			l := len(player.queue)
			logger.Debug("Queue RTP",
				log.String("player", player.String()),
				log.Uint("exceeds limit", player.queueLimit),
				log.Int("dropped old packets", oldLen-l),
				log.Int("current queue len", l),
			)
		}
	}
	player.cond.Signal()
	player.cond.L.Unlock()
	return player
}

func (player *Player) Start() {
	logger := player.logger
	timer := time.Unix(0, 0)
	for !player.Stopped {
		var pack *RTPPack
		player.cond.L.Lock()
		if len(player.queue) == 0 {
			player.cond.Wait()
		}
		if len(player.queue) > 0 {
			pack = player.queue[0]
			player.queue = player.queue[1:]
		}
		queueLen := len(player.queue)
		player.cond.L.Unlock()
		if player.paused {
			continue
		}
		if pack == nil {
			if !player.Stopped {
				logger.Warn("player not stoped, but queue take out nil pack")
			}
			continue
		}
		if err := player.SendRTP(pack); err != nil {
			logger.ErrorWith("rtsp player send rtp error", err)
		}
		elapsed := time.Now().Sub(timer)
		if config.RtspConfig().EnableDebug && elapsed >= 30*time.Second {
			logger.Debug("Send RTP",
				log.String("player", player.String()),
				log.String("package type", pack.Type.String()),
				log.Int("queue len", queueLen),
			)
			timer = time.Now()
		}
	}
}

func (player *Player) Pause(paused bool) {
	if paused {
		player.logger.Info("Player Pause", log.String("player", player.String()))
	} else {
		player.logger.Info("Player Play", log.String("player", player.String()))
	}
	player.cond.L.Lock()
	if paused && player.dropPacketWhenPaused && len(player.queue) > 0 {
		player.queue = make([]*RTPPack, 0)
	}
	player.paused = paused
	player.cond.L.Unlock()
}
