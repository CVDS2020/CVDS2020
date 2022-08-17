package main

import (
	"context"
	"fmt"
	"github.com/CVDS2020/CVDS2020/common/assert"
	"github.com/CVDS2020/CVDS2020/common/log"
	"github.com/CVDS2020/CVDS2020/cvds-mdu/args"
	"github.com/CVDS2020/CVDS2020/cvds-mdu/config"
	"github.com/CVDS2020/CVDS2020/cvds-mdu/routers"
	"github.com/CVDS2020/CVDS2020/cvds-mdu/rtsp"
	"github.com/CVDS2020/CVDS2020/cvds-mdu/system/service"
	"github.com/CVDS2020/CVDS2020/cvds-mdu/utils"
	"github.com/common-nighthawk/go-figure"
	"net/http"
	"strings"
	"time"
)

var Logger = assert.Must(config.LogConfig().Build("main"))

type program struct {
	httpServer *http.Server
	rtspServer *rtsp.Server
}

func (p *program) StopHTTP() (err error) {
	if p.httpServer == nil {
		err = fmt.Errorf("HTTP Server Not Found")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = p.httpServer.Shutdown(ctx); err != nil {
		return
	}
	return
}

func (p *program) StartHTTP() (err error) {
	p.httpServer = &http.Server{
		Addr:              config.HttpConfig().GetAddr(),
		Handler:           routers.Router,
		ReadTimeout:       config.HttpConfig().ReadTimeout,
		ReadHeaderTimeout: config.HttpConfig().ReadHeaderTimeout,
		WriteTimeout:      config.HttpConfig().WriteTimeout,
		IdleTimeout:       config.HttpConfig().IdleTimeout,
	}
	Logger.Info(fmt.Sprintf("http server start --> http://%s", p.httpServer.Addr))
	go func() {
		if err := p.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			Logger.ErrorWith("start http server error", err)
		}
		Logger.Info("http server end")
	}()
	return
}

func (p *program) StartRTSP() {
	if p.rtspServer == nil {
		Logger.Fatal("RTSP Server Not Found")
	}
	addr := p.rtspServer.Addr()
	Logger.Info(fmt.Sprintf("rtsp server start --> rtsp://%s:%d", addr.IP.String(), addr.Port))
	go func() {
		if err := p.rtspServer.Start(); err != nil {
			Logger.ErrorWith("start rtsp server error", err)
		}
		Logger.Info("rtsp server end")
	}()
	return
}

func (p *program) StopRTSP() (err error) {
	if p.rtspServer == nil {
		Logger.Fatal("RTSP Server Not Found")
	}
	p.rtspServer.Stop()
	return
}

func (p *program) Start(s service.Service) (err error) {
	Logger.Info("********** START **********")
	err = routers.Init()
	if err != nil {
		return
	}
	p.StartRTSP()
	p.StartHTTP()

	go func() {
		for range routers.API.RestartChan {
			p.StopHTTP()
			p.StopRTSP()
			config.ReloadConfig()
			p.StartRTSP()
			p.StartHTTP()
		}
	}()

	return
}

func (p *program) Stop(s service.Service) (err error) {
	defer Logger.Info("********** STOP **********")
	p.StopHTTP()
	p.StopRTSP()
	return
}

func printFigure() {
	defer func() {
		if e := recover(); e != nil {
			if err, is := e.(error); is {
				Logger.Fatal("print figure error", log.Error(err))
			} else {
				Logger.Fatal(fmt.Sprintf("print figure error: %v", e))
			}
		}
	}()
	figureConfig := config.FigureConfig()
	if figureConfig.Color == "" {
		figure.NewFigure(figureConfig.Phrase, figureConfig.Font, figureConfig.Strict).Print()
	} else {
		figure.NewColorFigure(figureConfig.Phrase, figureConfig.Font, figureConfig.Color, figureConfig.Strict).Print()
	}
}

func main() {
	svcConfig := &service.Config{
		Name:        config.ServiceConfig().Name,
		DisplayName: config.ServiceConfig().DisplayName,
		Description: config.ServiceConfig().Description,
	}

	rtspServer := rtsp.GetServer()
	p := &program{
		rtspServer: rtspServer,
	}
	s, err := service.New(p, svcConfig)
	if err != nil {
		Logger.ErrorWith("create service error", err)
		utils.PauseExit()
	}
	cmd := strings.ToLower(args.GetArgsConfig().Command)
	if cmd == "install" || cmd == "stop" || cmd == "start" || cmd == "uninstall" {
		printFigure()
		Logger.Info(fmt.Sprintf("%s %s ...", svcConfig.Name, cmd))
		if err = service.Control(s, cmd); err != nil {
			Logger.ErrorWith(fmt.Sprintf("%s %s failed", svcConfig.Name, cmd), err)
			utils.PauseExit()
		}
		Logger.Info(fmt.Sprintf("%s %s ok", svcConfig.Name, cmd))
		return
	}
	printFigure()
	if err = s.Run(); err != nil {
		Logger.ErrorWith("service run error", err)
		utils.PauseExit()
	}
}
