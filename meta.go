package coze

import (
	"net/http"

	"github.com/fimreal/rack/module"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	ID            = "coze"
	Comment       = "easy to share coze bots/workflow"
	RoutePrefix   = "/"
	DefaultEnable = false
)

var Module = module.Module{
	ID:          ID,
	Comment:     Comment,
	RouteFunc:   AddRoute,
	RoutePrefix: RoutePrefix,
	FlagFunc:    ServeFlag,
}

func ServeFlag(serveCmd *cobra.Command) {
	serveCmd.Flags().Bool(ID, DefaultEnable, Comment)
	serveCmd.Flags().String(ID+"_token", "", "coze token")
	serveCmd.Flags().String(ID+"_url", "https://api.coze.cn", "coze api url")
}

func AddRoute(g *gin.Engine) {
	if !viper.GetBool(ID) && !viper.GetBool("allservices") {
		return
	}
	g.GET("/help/"+ID, help)

	r := g.Group(RoutePrefix)
	r.POST("/txt2img", handleTxt2Img)
}

func help(ctx *gin.Context) {
	ctx.String(http.StatusOK, `POST /txt2imgcurl
--header 'Content-Type: application/json'
--data-raw '{
    "user_id": "12345",              # 用户 ID, 按照您的需求替换成实际值
    "botid": "7396***",               # bot ID
    "prompt": "给我画一幅1:1动漫风格插画，落叶背景，可爱的少女",   # 您要传递给 Bot 的提示信息
    "conversation_id": ""            # 可选参数，如有上下文可填入
}'
`)
}
