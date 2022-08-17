package main

import (
	"context"
	"fmt"
	"github.com/CVDS2020/CVDS2020/common/assert"
	"github.com/CVDS2020/CVDS2020/common/log"
	"github.com/CVDS2020/CVDS2020/cvds-msu/args"
	"github.com/CVDS2020/CVDS2020/cvds-msu/config"
	"github.com/CVDS2020/CVDS2020/cvds-msu/controller"
	"github.com/CVDS2020/CVDS2020/cvds-msu/system/service"
	"github.com/CVDS2020/CVDS2020/cvds-msu/utils"
	"github.com/common-nighthawk/go-figure"
	"net/http"
	"strings"
	"time"
)

var Logger = assert.Must(config.LogConfig().Build("main"))

type program struct {
	httpServer *controller.Server
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
	p.httpServer.Wait()
	return
}

func (p *program) StartHTTP() (err error) {
	p.httpServer = controller.NewServer()
	Logger.Info(fmt.Sprintf("http server start --> http://%s", p.httpServer.Addr))
	go func() {
		if err := p.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			Logger.ErrorWith("start http server error", err)
		}
		Logger.Info("http server end")
	}()
	return
}

func (p *program) Start(s service.Service) (err error) {
	Logger.Info("********** START **********")
	if err != nil {
		return
	}
	p.StartHTTP()

	go func() {
		for range controller.API.RestartChan {
			p.StopHTTP()
			config.ReloadConfig()
			p.StartHTTP()
		}
	}()

	return
}

func (p *program) Stop(s service.Service) (err error) {
	defer Logger.Info("********** STOP **********")
	p.StopHTTP()
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

	p := &program{}
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
