package controller

import (
	"github.com/CVDS2020/CVDS2020/cvds-msu/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
)

type Channel struct {
	svc *service.Channel
}

func (s *Channel) StartChannel(ctx *gin.Context) {
	model := struct {
		Name      string         `json:"name"`
		URL       string         `json:"url"`
		Transport string         `json:"transport"`
		Cover     uint           `json:"cover"`
		Fields    map[string]any `yaml:"fields" json:"fields"`
	}{}
	if err := ctx.ShouldBindJSON(&model); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind)
		return
	}

	channel, err := s.svc.CreateChannel(model.Name, model.URL, model.Transport, model.Cover, model.Fields)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusBadRequest,
			"msg":  err.Error(),
		})
		ctx.Abort()
		return
	}

	if err = channel.Start(); err != nil {
		s.svc.RemoveChannel(channel.UUID())
		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusBadRequest,
			"msg":  err.Error(),
		})
		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "success",
		"channel": gin.H{
			"uuid": channel.UUID(),
		},
	})
}

func (s *Channel) GetChannel(ctx *gin.Context) {
	id := ctx.Query("uuid")
	channel, err := s.svc.GetChannel(id)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code": http.StatusBadRequest,
			"msg":  err.Error(),
		})
		ctx.Abort()
		return
	}
	model := gin.H{
		"uuid":      channel.UUID(),
		"name":      channel.Name(),
		"url":       channel.URL(),
		"transport": channel.Transport(),
		"cover":     channel.Cover(),
		"fields":    channel.Fields(),
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"msg":     "success",
		"channel": model,
	})
}

func (s *Channel) StopChannel(c *gin.Context) {
	id := c.Query("uuid")
	if err := s.svc.RemoveChannel(id); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": http.StatusBadRequest,
			"msg":  err.Error(),
		})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "success",
	})
}

func (s *Channel) StopAll() {
	s.svc.RemoveAll()
}

var channel *Channel
var channelInitializer = sync.Once{}

func GetChannel() *Channel {
	if channel != nil {
		return channel
	}
	channelInitializer.Do(func() {
		channel = &Channel{
			svc: service.GetChannel(),
		}
	})
	return GetChannel()
}
