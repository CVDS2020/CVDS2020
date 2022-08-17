package routers

import (
	"fmt"
	"github.com/CVDS2020/CVDS2020/common/assert"
	"github.com/CVDS2020/CVDS2020/common/def"
	"github.com/CVDS2020/CVDS2020/common/log"
	"github.com/CVDS2020/CVDS2020/cvds-mdu/config"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/validator.v8"
	"net/http"
	"time"
)

/**
 * @apiDefine simpleSuccess
 * @apiSuccessExample 成功
 * HTTP/1.1 200 OK
 */

/**
 * @apiDefine authError
 * @apiErrorExample 认证失败
 * HTTP/1.1 401 access denied
 */

/**
 * @apiDefine pageParam
 * @apiParam {Number} start 分页开始,从零开始
 * @apiParam {Number} limit 分页大小
 * @apiParam {String} [sort] 排序字段
 * @apiParam {String=ascending,descending} [order] 排序顺序
 * @apiParam {String} [q] 查询参数
 */

/**
 * @apiDefine pageSuccess
 * @apiSuccess (200) {Number} total 总数
 * @apiSuccess (200) {Array} rows 分页数据
 */

/**
 * @apiDefine timeInfo
 * @apiSuccess (200) {String} rows.createAt 创建时间, YYYY-MM-DD HH:mm:ss
 * @apiSuccess (200) {String} rows.updateAt 结束时间, YYYY-MM-DD HH:mm:ss
 */

var Router *gin.Engine
var Logger *log.Logger

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

func loggerFormatter(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}

	Logger.Debug(fmt.Sprintf("|%s %3d %s| %13v |%s %-7s %s|",
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		methodColor, param.Method, resetColor),
		log.String("client", param.ClientIP),
		log.String("path", param.Path),
	)

	return ""
}

func Init() (err error) {
	Router = gin.New()
	Logger = assert.Must(config.LogConfig().Build("routers")).WithOptions(log.WithCaller(false))
	pprof.Register(Router)

	Router.Use(gin.LoggerWithFormatter(loggerFormatter))
	Router.Use(gin.Recovery())
	Router.Use(Errors())

	{
		api := Router.Group("/api/v1")
		api.GET("/restart", API.Restart)

		api.GET("/pushers", API.Pushers)
		api.GET("/players", API.Players)

		api.GET("/stream/start", API.StreamStart)
		api.GET("/stream/stop", API.StreamStop)
	}

	return
}
