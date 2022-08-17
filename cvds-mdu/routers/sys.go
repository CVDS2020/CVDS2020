package routers

import (
	"github.com/gin-gonic/gin"
	"net/http"
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

// Restart
/* @api {get} /api/v1/restart 重启服务
 * @apiGroup sys
 * @apiName Restart
 * @apiUse simpleSuccess
 */
func (h *APIHandler) Restart(c *gin.Context) {
	Logger.Info("Restart...")
	c.JSON(http.StatusOK, "OK")
	go func() {
		select {
		case h.RestartChan <- true:
		default:
		}
	}()
}
