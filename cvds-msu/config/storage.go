package config

import (
	"github.com/CVDS2020/CVDS2020/common/config"
	"github.com/CVDS2020/CVDS2020/common/uns/goos"
	"time"
)

type Storage struct {
	DataDir string `yaml:"data-dir" json:"data-dir"`
	TmpDir  string `yaml:"tmp-dir" json:"tmp-dir"`
	FFMpeg  struct {
		Bin                 string        `yaml:"bin" json:"bin"`
		ExitRestartInterval time.Duration `yaml:"exit-restart-interval" json:"exit-restart-interval"`
		InterruptTimeout    time.Duration `yaml:"interrupt-timeout" json:"interrupt-timeout"`
	}
	FileName     string `yaml:"file-name" json:"file-name"`
	FileDuration uint   `yaml:"file-duration" json:"file-duration"`
	FileFormat   string `yaml:"file-format" json:"file-format"`
	TimeLayout   string `yaml:"time-layout" json:"time-layout"`
	//SeqLayout     string `yaml:"seq-layout" json:"seq-layout"`
	MoveInterval        uint `yaml:"move-interval" json:"move-interval"`
	CheckDeleteInterval uint `yaml:"check-delete-interval" json:"check-delete-interval"`
}

func (s *Storage) PreHandle() config.PreHandlerConfig {
	s.DataDir = "data"
	switch goos.GOOS {
	case "windows":
		s.FFMpeg.Bin = "ffmpeg.exe"
	default:
		s.FFMpeg.Bin = "ffmpeg"
	}
	s.FFMpeg.ExitRestartInterval = time.Second
	s.FFMpeg.InterruptTimeout = time.Second
	s.FileName = "{{.channel}}.{{.suffix}}"
	s.FileDuration = 10 * 60
	s.FileFormat = "mp4"
	s.TimeLayout = "2006-01-02_15h04m05s"
	//s.SeqLayout = "%05d"
	s.MoveInterval = 2
	s.CheckDeleteInterval = 2
	return s
}
