package web

import (
	"fmt"
	"log"
	"os"

	"nxtrace-server/server/common"

	"github.com/gin-gonic/gin"
)

type TraceData struct {
	Region string
	Host   string
	Params []string
}

func trace_hanele(c *gin.Context) {
	var traceData TraceData

	err := c.ShouldBindJSON(&traceData)
	if err != nil {
		c.String(503, err.Error())
		return
	}

	host := traceData.Host
	params := traceData.Params

	output, err := common.RunTrace(host, params)
	if err != nil {
		c.String(503, err.Error())
		return
	}

	c.String(200, output)
}

func Run() error {
	config := new(common.Config)
	webCfg := config.NewWebConfig()

	if os.Getenv("DEBUG") == "true" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	listenAddr := fmt.Sprintf("%s:%s", webCfg.ServerHost, webCfg.ServerPort)
	log.Printf("Listenning: %s\n", listenAddr)

	router := gin.New()
	router.Use(gin.Recovery(), gin.Logger())
	router.SetTrustedProxies(nil)

	router.POST("/trace", trace_hanele)
	return router.Run(listenAddr)
}
