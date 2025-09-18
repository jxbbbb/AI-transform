package main

import (
	"ai-transform-backend/pkg/config"
	"ai-transform-backend/pkg/log"
	"ai-transform-backend/pkg/storage/cos"
	"ai-transform-backend/transform-refer-api/controllers"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
)

var (
	configFile = flag.String("config", "dev.referapi.config.yaml", "")
)

func main() {
	flag.Parse()
	//初始化配置文件
	config.InitConfig(*configFile)
	conf := config.GetConfig()
	//初始化日志组件
	log.SetLevel(conf.Log.Level)
	log.SetOutput(log.GetRotateWriter(conf.Log.LogPath))
	log.SetPrintCaller(true)

	//初始化日志组件
	logger := log.NewLogger()
	logger.SetOutput(log.GetRotateWriter(conf.Log.LogPath))
	logger.SetLevel(conf.Log.Level)
	logger.SetPrintCaller(true)

	csf := cos.NewCosStorageFactory(
		conf.Cos.BucketUrl,
		conf.Cos.SecretId,
		conf.Cos.SecretKey,
		conf.Cos.CDNDomain)

	controller := controllers.NewReferWav(csf, conf, logger)

	r := gin.Default()
	r.GET("/health", func(*gin.Context) {})
	api := r.Group("/api")
	api.POST("/refer/wav", controller.SaveReferWav)
	api.GET("/refer/wav", controller.GetReferInfo)

	err := r.Run(fmt.Sprintf("%s:%d", conf.Http.IP, conf.Http.Port))
	if err != nil {
		log.Fatal(err)
	}
}
