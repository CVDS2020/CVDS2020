package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
)

/**
 * @apiDefine sys 系统
 */

type APIHandler struct {
	RestartChan chan bool
}

var API = &APIHandler{
	RestartChan: make(chan bool),
}

type Sys struct {
	RestartChan chan bool
}

// Restart
/* @api {post} /api/v1/restart 重启服务
 * @apiGroup sys
 * @apiName Restart
 * @apiUse simpleSuccess
 */
func (c *Sys) Restart(ctx *gin.Context) {
	Logger.Info("Restart...")
	ctx.JSON(http.StatusOK, "OK")
	go func() {
		select {
		case c.RestartChan <- true:
		default:
		}
	}()
}

var sys *Sys
var sysInitializer = sync.Once{}

func GetSys() *Sys {
	if sys != nil {
		return sys
	}
	sysInitializer.Do(func() {
		sys = &Sys{RestartChan: make(chan bool)}
	})
	return sys
}
