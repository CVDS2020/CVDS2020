package storage

import (
	"context"
	"github.com/CVDS2020/CVDS2020/common/assert"
	"github.com/CVDS2020/CVDS2020/common/errors"
	"github.com/CVDS2020/CVDS2020/common/lifecycle"
	"github.com/CVDS2020/CVDS2020/common/log"
	"github.com/CVDS2020/CVDS2020/common/timer"
	"github.com/CVDS2020/CVDS2020/cvds-msu/config"
	"github.com/CVDS2020/CVDS2020/cvds-msu/utils"
	"github.com/gofrs/uuid"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var (
	InvalidChannelNameError = errors.New("invalid channel name")
	InvalidChannelURLError  = errors.New("invalid channel url")
)

var (
	ChannelExistError      = errors.New("channel exist")
	ChannelNotFoundError   = errors.New("channel not found")
	ChannelDestroyedError  = errors.New("channel destroyed")
	ChannelClosedError     = errors.New("channel has been closed")
	ChannelIsRunningError  = errors.New("channel is running")
	ChannelRestartingError = errors.New("channel is restarting")
)

type ChannelState struct {
	Closed     bool
	Running    bool
	Closing    bool
	Restarting bool
}

type Channel struct {
	lifecycle.Lifecycle
	runner       *lifecycle.DefaultRunner
	notAvailable bool

	uuid      string
	name      string
	url       string
	transport string
	cover     uint
	fields    map[string]any

	//seq int64

	fileNameTemp *template.Template
	fileDuration uint
	fileFormat   string
	timeLayout   string
	//seqLayout    string
	dataDir string
	tmpDir  string

	destroyRequest bool
	destroyed      bool

	closeSignal chan struct{}
	logger      *log.Logger
}

func NewChannel(name string, url string, transport string, cover uint, fields map[string]any) *Channel {
	c := new(Channel)
	c.init(name, url, transport, cover, fields)
	return c
}

func (c *Channel) init(name string, url string, transport string, cover uint, fields map[string]any) {
	c.uuid = uuid.Must(uuid.NewV4()).String()
	c.name = name
	c.url = url
	c.transport = transport
	c.cover = cover
	c.fields = fields
	c.closeSignal = make(chan struct{}, 1)
	//c.seq = -1
	c.logger = assert.Must(config.LogConfig().Build("storage.channel"))
	c.runner, c.Lifecycle = lifecycle.New("channel", c.doStart, c.doRun, c.doClose,
		lifecycle.StartChecker(c.startChecker),
	)
}

func (c *Channel) moveTmpToData() error {
	fileInfos, err := ioutil.ReadDir(c.tmpDir)
	if err != nil {
		return c.logger.ErrorWith("list tmp directory error", err, log.String("tmp directory", c.tmpDir))
	}

	//var removeList []struct {
	//	seq  int64
	//	path string
	//}

	var moveList []struct {
		createTime time.Time
		src        string
		target     string
		targetDir  string
	}

	//maxSeq := c.seq
	maxTime := time.Time{}
	for _, info := range fileInfos {
		name := info.Name()
		filePath := path.Join(c.tmpDir, info.Name())
		if strings.HasSuffix(name, "."+c.fileFormat) {
			createTime, err := time.Parse("20060102150405", name[:len(name)-4])
			if err != nil {
				// ignore invalid format
				c.logger.Debug("invalid file name format, ignored", log.String("file", info.Name()))
				continue
			}

			if createTime.After(maxTime) {
				maxTime = createTime
			}

			//seq, err := strconv.ParseInt(name[:len(name)-4], 10, 64)
			//if err != nil || seq < 0 {
			//	// ignore invalid seq
			//	c.logger.Debug("invalid file name format, ignored", log.String("file", info.Name()))
			//	continue
			//}
			//createTime = utils.GetFileCreateTime(info)
			//
			//// check seq
			//if seq < c.seq {
			//	removeList = append(removeList, struct {
			//		seq  int64
			//		path string
			//	}{seq: seq, path: filePath})
			//	continue
			//}
			//
			//// check expired and add to remove list
			//if time.Now().Sub(createTime) > time.Duration(c.cover)*time.Hour {
			//	removeList = append(removeList, struct {
			//		seq  int64
			//		path string
			//	}{seq: seq, path: filePath})
			//	continue
			//}
			//
			//if seq > maxSeq {
			//	maxSeq = seq
			//}

			// generate file name
			ctx := make(map[string]any, len(c.fields))
			for k, v := range c.fields {
				ctx[k] = v
			}
			ctx["channel"] = c.name
			ctx["suffix"] = c.fileFormat
			sb := &strings.Builder{}
			sb.WriteString(createTime.Format(c.timeLayout))
			sb.WriteByte('_')
			c.fileNameTemp.Execute(sb, ctx)
			fileName := sb.String()

			// generate target directory name
			targetDir := createTime.Format("2006-01-02")
			// generate target path
			target := path.Join(c.dataDir, targetDir, fileName)
			// add to move list
			moveList = append(moveList, struct {
				createTime time.Time
				src        string
				target     string
				targetDir  string
			}{createTime: createTime, src: filePath, target: target, targetDir: targetDir})
		} else {
			// ignore
			c.logger.Debug("invalid file name format, ignored", log.String("file", info.Name()))
		}
	}

	//for _, entry := range removeList {
	//	if entry.seq < maxSeq {
	//		if err := os.Remove(entry.path); err != nil {
	//			c.logger.ErrorWith("remove file error", err, log.String("path", entry.path))
	//		} else {
	//			c.logger.Info("remove file success", log.String("path", entry.path))
	//		}
	//	} else {
	//		c.logger.Debug("ignore remove latest sequence file", log.String("path", entry.path))
	//	}
	//}

	for _, entry := range moveList {
		if entry.createTime.Before(maxTime) {
			targetDirPath := path.Join(c.dataDir, entry.targetDir)
			if err := utils.EnsureDir(targetDirPath); err != nil {
				c.logger.ErrorWith("ensure target directory error", err, log.String("target directory", targetDirPath))
				continue
			}
			if err := os.Rename(entry.src, entry.target); err != nil {
				c.logger.ErrorWith("move file error", err, log.String("src", entry.src), log.String("target", entry.target))
			} else {
				c.logger.Info("move file success", log.String("src", entry.src), log.String("target", entry.target))
			}
		} else {
			c.logger.Debug("ignore move latest time file", log.String("src", entry.src))
		}
	}

	return nil
}

func (c *Channel) checkAndDelete() error {
	dirInfos, err := ioutil.ReadDir(c.dataDir)
	if err != nil {
		return c.logger.ErrorWith("list data directory error", err, log.String("data directory", c.dataDir))
	}

	minDate := time.Time{}
	minDataDir := ""
	for _, info := range dirInfos {
		if info.Name() == ".tmp" {
			continue
		}

		date, err := time.Parse("2006-01-02", info.Name())
		if err != nil {
			// ignore
			c.logger.Debug("invalid directory name format, ignored", log.String("directory", info.Name()))
			continue
		}

		if minDate.IsZero() || date.Before(minDate) {
			minDate = date
			minDataDir = info.Name()
		}
	}

	if minDataDir == "" {
		c.logger.Info("channel data directory not found")
		return nil
	}

	dirPath := path.Join(c.dataDir, minDataDir)
	fileInfos, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return c.logger.ErrorWith("list channel data directory error", err, log.String("data directory", dirPath))
	}

	var removeList []struct {
		path string
	}

	for _, info := range fileInfos {
		name := info.Name()
		filePath := path.Join(dirPath, info.Name())

		createTime, err := time.ParseInLocation(c.timeLayout, name[:len(c.timeLayout)], time.Local)
		if err != nil {
			// ignore
			c.logger.Debug("invalid file create time format, ignored", log.String("file", name), log.Error(err))
			continue
		}

		if time.Now().Sub(createTime) > time.Duration(c.cover)*time.Minute {
			removeList = append(removeList, struct {
				path string
			}{path: filePath})
			continue
		}
	}

	if len(removeList) != 0 {
		c.logger.Debug("has expired file in channel data directory", log.Int("expired file count", len(removeList)))
	} else {
		c.logger.Debug("expired file not found in channel data directory")
	}

	for _, entry := range removeList {
		if err := os.Remove(entry.path); err != nil {
			c.logger.ErrorWith("remove file error", err, log.String("path", entry.path))
		} else {
			c.logger.Info("remove file success", log.String("path", entry.path))
		}
	}

	return nil
}

func (c *Channel) doStart() error {
	storageConfig := config.StorageConfig()
	fileDuration := storageConfig.FileDuration
	if fileDuration < 10 {
		fileDuration = 10
	}
	// create file name template
	t, err := template.New("").Parse(storageConfig.FileName)
	if err != nil {
		return c.logger.ErrorWith("parse file name template error", err, log.String("file name template", storageConfig.FileFormat))
	}

	// check and ensure root data directory
	rootDataDir := storageConfig.DataDir
	if err := utils.EnsureDir(rootDataDir); err != nil {
		return c.logger.ErrorWith("ensure root data directory error", err, log.String("root data directory", rootDataDir))
	}

	// check and ensure data directory
	dataDir := path.Join(rootDataDir, c.name)
	if err := utils.EnsureDir(dataDir); err != nil {
		return c.logger.ErrorWith("ensure data directory error", err, log.String("data directory", dataDir))
	}

	// check and ensure tmp directory
	tmpDir := path.Join(dataDir, ".tmp")
	if err := utils.EnsureDir(tmpDir); err != nil {
		return c.logger.ErrorWith("ensure tmp directory error", err, log.String("data directory", tmpDir))
	}

	c.fileNameTemp = t
	c.fileDuration = fileDuration
	c.fileFormat = storageConfig.FileFormat
	c.timeLayout = storageConfig.TimeLayout
	//c.timeLayout, c.seqLayout = storageConfig.TimeLayout, storageConfig.SeqLayout
	c.dataDir, c.tmpDir = dataDir, tmpDir

	c.moveTmpToData()
	return nil
}

func (c *Channel) doRun() error {
	//moveRequestChan := make(chan struct{}, 1)
	//moveResponseChan := make(chan struct{}, 1)

	ffmpegSignal := make(chan os.Signal, 1)
	moverStopChan := make(chan struct{}, 1)
	deleterStopChan := make(chan struct{}, 1)

	ffmpegStoppedCtx, ffmpegStopped := context.WithCancel(context.Background())
	moverStoppedCtx, moverStopped := context.WithCancel(context.Background())
	deleterStoppedCtx, deleterStopped := context.WithCancel(context.Background())

	// move file from tmp to data directory goroutine
	go func() {
		c.logger.Info("move file process start")
		defer c.logger.Info("move file process stopped")
		defer moverStopped()

		checkInterval := time.Duration(config.StorageConfig().MoveInterval) * time.Second
		moverTimer := timer.NewTimer(make(chan struct{}, 1))
		defer moverTimer.Stop()

		for moverTimer.After(checkInterval); true; {
			select {
			case <-moverTimer.C:
				c.moveTmpToData()
				moverTimer.After(checkInterval)
			//case <-moveRequestChan:
			//	c.moveTmpToData()
			//	moveResponseChan <- struct{}{}
			//	moverTimer.After(checkInterval)
			case <-moverStopChan:
				return
			}
		}
	}()

	// delete expired file goroutine
	go func() {
		c.logger.Info("delete expired file process start")
		defer c.logger.Info("check delete process stopped")
		defer deleterStopped()

		checkDeleteInterval := time.Duration(config.StorageConfig().CheckDeleteInterval) * time.Second
		deleterTimer := timer.NewTimer(make(chan struct{}, 1))
		defer deleterTimer.Stop()

		for deleterTimer.After(checkDeleteInterval); true; {
			select {
			case <-deleterTimer.C:
				c.checkAndDelete()
				deleterTimer.After(checkDeleteInterval)
			case <-deleterStopChan:
				return
			}
		}
	}()

	// ffmpeg goroutine
	go func() {
		c.logger.Info("ffmpeg process start")
		defer c.logger.Info("ffmpeg process stopped")
		defer ffmpegStopped()

		ffmpegBin := config.StorageConfig().FFMpeg.Bin
		restartTimer := timer.NewTimer(make(chan struct{}, 1))
		defer restartTimer.Stop()

		exitChan, killChan := make(chan error, 1), make(chan struct{}, 1)
		closing, started := false, false

		var cmd *exec.Cmd
		for restartTimer.Trigger(); true; {
			select {
			case <-restartTimer.C:
				//moveRequestChan <- struct{}{}
				//<-moveResponseChan
				args := []string{
					"-rtsp_transport", c.transport,
					"-i", c.url,
					"-codec", "copy",
					"-f", "segment",
					"-strftime", "1",
					"-segment_time", strconv.FormatUint(uint64(c.fileDuration), 10),
					"-segment_format", c.fileFormat,
					//"-segment_start_number", strconv.FormatUint(uint64(c.seq+1), 10),
					path.Join(c.tmpDir, "%Y%m%d%H%M%S."+c.fileFormat),
				}
				cmd = exec.Command(ffmpegBin, args...)
				// start error retry or exit retry
				if err := cmd.Start(); err != nil {
					c.logger.ErrorWith("ffmpeg start error", err, log.String("cmd", strings.Join(append([]string{ffmpegBin}, args...), " ")))
					restartTimer.After(config.StorageConfig().FFMpeg.ExitRestartInterval)
					continue
				}
				c.logger.Info("ffmpeg started", log.String("cmd", strings.Join(append([]string{ffmpegBin}, args...), " ")))
				// cmd started
				started = true
				go func() {
					exitChan <- cmd.Wait()
				}()

			case err := <-exitChan:
				// cmd run completion
				started = false
				if err != nil {
					c.logger.ErrorWith("ffmpeg exit error", err)
				} else {
					c.logger.Warn("ffmpeg exit")
				}
				if closing {
					return
				}
				restartTimer.After(config.StorageConfig().FFMpeg.ExitRestartInterval)

			case signal := <-ffmpegSignal:
				closing = true
				if !started || cmd == nil {
					// cmd run completion
					return
				}
				if process := cmd.Process; process != nil {
					c.logger.Info("request interrupt ffmpeg")
					process.Signal(signal)
					time.AfterFunc(config.StorageConfig().FFMpeg.InterruptTimeout, func() {
						killChan <- struct{}{}
					})
					continue
				}
				return

			case <-killChan:
				if !started {
					// cmd run completion
					return
				}
				if process := cmd.Process; process != nil {
					c.logger.Info("interrupt timeout, kill ffmpeg")
					process.Kill()
					continue
				}
				return
			}
		}
	}()

	// request close
	<-c.closeSignal

	// stop ffmpeg
	ffmpegSignal <- os.Interrupt
	// stop deleter
	deleterStopChan <- struct{}{}
	// stop mover
	moverStopChan <- struct{}{}

	<-ffmpegStoppedCtx.Done()
	<-deleterStoppedCtx.Done()
	<-moverStoppedCtx.Done()

	if c.destroyRequest {
		c.destroyed = true
	}

	return nil
}

func (c *Channel) doClose() error {
	c.closeSignal <- struct{}{}
	return nil
}

func (c *Channel) startChecker() error {
	if c.destroyed {
		return ChannelDestroyedError
	}
	return nil
}

// Start wrap Lifecycle.Start, covert lifecycle.StateError to localize error
func (c *Channel) Start() error {
	if err := c.Lifecycle.Start(); err != nil {
		if se, ok := err.(lifecycle.StateError); ok {
			switch se.TypeId {
			case lifecycle.StateRunningErrorTypeId:
				return ChannelIsRunningError
			case lifecycle.StateRestartingErrorTypeId:
				return ChannelRestartingError
			}
		}
		return err
	}
	return nil
}

// Close wrap Lifecycle.Close, covert lifecycle.StateError to localize error
func (c *Channel) Close(future chan error) error {
	if err := c.Lifecycle.Close(future); err != nil {
		if se, ok := err.(lifecycle.StateError); ok {
			switch se.TypeId {
			case lifecycle.StateClosedErrorTypeId:
				return ChannelClosedError
			case lifecycle.StateRestartingErrorTypeId:
				return ChannelRestartingError
			}
		}
		return err
	}
	return nil
}

// UUID function is getter of Channel.uuid
func (c *Channel) UUID() string {
	return c.uuid
}

// Name function is getter of Channel.name
func (c *Channel) Name() string {
	return c.name
}

// URL function is getter of Channel.url
func (c *Channel) URL() string {
	return c.url
}

// Transport function is getter of Channel.transport
func (c *Channel) Transport() string {
	return c.transport
}

func (c *Channel) Cover() uint {
	return c.cover
}

func (c *Channel) Fields() map[string]any {
	return c.fields
}

func (c *Channel) Destroy() error {
	c.destroyRequest = true
	err, _ := c.CloseWait()
	return err
}
