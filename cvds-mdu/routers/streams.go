package routers

import (
	"fmt"
	"github.com/CVDS2020/CVDS2020/common/log"
	"github.com/CVDS2020/CVDS2020/cvds-mdu/config"
	"github.com/CVDS2020/CVDS2020/cvds-mdu/rtsp"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

/**
 * @apiDefine stream 流管理
 */

// StreamStart
/* @api {get} /api/v1/stream/start 启动拉转推
 * @apiGroup stream
 * @apiName StreamStart
 * @apiParam {String} url RTSP源地址
 * @apiParam {String} [customPath] 转推时的推送PATH
 * @apiParam {String=TCP,UDP} [transType=TCP] 拉流传输模式
 * @apiParam {Number} [idleTimeout] 拉流时的超时时间
 * @apiParam {Number} [heartbeatInterval] 拉流时的心跳间隔，毫秒为单位。如果心跳间隔不为0，那拉流时会向源地址以该间隔发送OPTION请求用来心跳保活
 * @apiSuccess (200) {String} ID	拉流的ID。后续可以通过该ID来停止拉流
 */
func (h *APIHandler) StreamStart(c *gin.Context) {
	type Form struct {
		URL               string `form:"url" binding:"required"`
		CustomPath        string `form:"customPath"`
		TransType         string `form:"transType"`
		IdleTimeout       int    `form:"idleTimeout"`
		HeartbeatInterval int    `form:"heartbeatInterval"`
	}
	var form Form
	err := c.Bind(&form)
	if err != nil {
		Logger.ErrorWith("Pull to push err:%v", err)
		return
	}
	agent := fmt.Sprintf("MDU/%s", config.GlobalConfig().Version)
	client, err := rtsp.NewRTSPClient(rtsp.GetServer(), form.URL, int64(form.HeartbeatInterval)*1000, agent)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	if form.CustomPath != "" && !strings.HasPrefix(form.CustomPath, "/") {
		form.CustomPath = "/" + form.CustomPath
	}
	client.CustomPath = form.CustomPath
	switch strings.ToLower(form.TransType) {
	case "udp":
		client.TransType = rtsp.TransTypeUdp
	case "tcp":
		fallthrough
	default:
		client.TransType = rtsp.TransTypeTcp
	}

	pusher := rtsp.NewClientPusher(client)
	if rtsp.GetServer().GetPusher(pusher.Path()) != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("Path %s already exists", client.Path))
		return
	}
	err = client.Start(time.Duration(form.IdleTimeout) * time.Second)
	if err != nil {
		Logger.ErrorWith("Pull stream error", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("Pull stream err: %v", err))
		return
	}
	rtsp.GetServer().AddPusher(pusher)
	Logger.Info("Pull to pusher success", log.String("pusher", pusher.String()))
	c.IndentedJSON(200, pusher.ID())
}

// StreamStop
/* @api {get} /api/v1/stream/stop 停止推流
 * @apiGroup stream
 * @apiName StreamStop
 * @apiParam {String} id 拉流的ID
 * @apiUse simpleSuccess
 */
func (h *APIHandler) StreamStop(c *gin.Context) {
	type Form struct {
		ID string `form:"id" binding:"required"`
	}
	var form Form
	err := c.Bind(&form)
	if err != nil {
		Logger.ErrorWith("stop pull to push error", err)
		return
	}
	pushers := rtsp.GetServer().GetPushers()
	for _, v := range pushers {
		if v.ID() == form.ID {
			v.Stop()
			c.IndentedJSON(200, "OK")
			Logger.Info("Stop pusher success", log.String("pusher", v.String()))
			return
		}
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("Pusher[%s] not found", form.ID))
}
