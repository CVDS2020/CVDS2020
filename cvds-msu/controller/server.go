package controller

import (
	"fmt"
	"github.com/CVDS2020/CVDS2020/common/assert"
	"github.com/CVDS2020/CVDS2020/common/def"
	"github.com/CVDS2020/CVDS2020/common/log"
	"github.com/CVDS2020/CVDS2020/cvds-msu/config"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/validator.v8"
	"net/http"
	"time"
)

var Logger = assert.Must(config.LogConfig().Build("controller"))

func init() {
	if !config.HttpConfig().Gin.EnableConsoleColor {
		gin.DisableConsoleColor()
	}
	gin.SetMode(gin.ReleaseMode)
}

func Errors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		for _, err := range c.Errors {
			switch err.Type {
			case gin.ErrorTypeBind:
				switch err.Err.(type) {
				case validator.ValidationErrors:
					errs := err.Err.(validator.ValidationErrors)
					for _, err := range errs {
						sec := config.GlobalConfig().Localize
						field := def.Default(sec[err.Field], err.Field)
						tag := def.Default(sec[err.Tag], err.Tag)
						c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("%s %s", field, tag))
						return
					}
				default:
					Logger.ErrorWith("", err.Err)
					c.AbortWithStatusJSON(http.StatusBadRequest, "Inner Error")
					return
				}
			}
		}
	}
}

type Server struct {
	*http.Server
	sys          *Sys
	channel      *Channel
	logger       *log.Logger
	closedFuture chan struct{}
}

func NewServer() *Server {
	s := new(Server)
	s.init()
	return s
}

func (s *Server) init() {
	s.Server = &http.Server{
		Addr:              config.HttpConfig().GetAddr(),
		ReadTimeout:       config.HttpConfig().ReadTimeout,
		ReadHeaderTimeout: config.HttpConfig().ReadHeaderTimeout,
		WriteTimeout:      config.HttpConfig().WriteTimeout,
		IdleTimeout:       config.HttpConfig().IdleTimeout,
	}

	s.logger = assert.Must(config.LogConfig().Build("controller.server")).WithOptions(log.WithCaller(false))
	s.closedFuture = make(chan struct{}, 1)

	router := gin.New()
	pprof.Register(router)
	router.Use(gin.LoggerWithFormatter(s.loggerFormatter))
	router.Use(gin.Recovery())
	router.Use(Errors())

	s.sys = GetSys()
	s.channel = GetChannel()

	api := router.Group("/api/v1")
	{
		sysApi := api.Group("/sys")
		{
			sysApi.POST("/restart", s.sys.Restart)
		}
		channelApi := api.Group("/channel")
		{
			channelApi.POST("/start", s.channel.StartChannel)
			channelApi.Group("/", s.channel.GetChannel)
			channelApi.DELETE("/stop", s.channel.StopChannel)
		}
	}

	s.Server.Handler = router
	s.Server.RegisterOnShutdown(s.OnShutdown)
}

func (s *Server) loggerFormatter(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}

	s.logger.Debug(fmt.Sprintf("|%s %3d %s| %13v |%s %-7s %s|",
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		methodColor, param.Method, resetColor),
		log.String("client", param.ClientIP),
		log.String("path", param.Path),
	)

	return ""
}

func (s *Server) Logger() *log.Logger {
	return s.logger
}

func (s *Server) OnShutdown() {
	s.channel.StopAll()
	s.closedFuture <- struct{}{}
}

func (s *Server) Wait() {
	<-s.closedFuture
}

//var server *Server
//var serverInitializer sync.Once
//
//func GetServer() *Server {
//	if server != nil {
//		return server
//	}
//	serverInitializer.Do(func() {
//		server = NewServer()
//	})
//	return GetServer()
//}
