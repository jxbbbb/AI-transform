package main

import (
	data2 "ai-transform-backend/data"
	"ai-transform-backend/pkg/config"
	"ai-transform-backend/pkg/db/mysql"
	"ai-transform-backend/pkg/log"
	"ai-transform-backend/pkg/mq/kafka"
	"ai-transform-backend/transform-web-api/controller"
	"ai-transform-backend/transform-web-api/middleware"
	"ai-transform-backend/transform-web-api/routers"
	"flag"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime"
)

var (
	configFile = flag.String("config", "dev.webapi.config.yaml", "")
)

func main() {
	flag.Parse()
	config.InitConfig(*configFile)
	cnf := config.GetConfig()

	log.SetLevel(cnf.Log.Level)
	log.SetOutput(log.GetRotateWriter(cnf.Log.LogPath))
	log.SetPrintCaller(true)
	logger := log.NewLogger()
	logger.SetLevel(cnf.Log.Level)
	logger.SetOutput(log.GetRotateWriter(cnf.Log.LogPath))
	logger.SetPrintCaller(true)

	kafka.InitKafkaProducerPool(kafka.ProducerPoolExternalKey, cnf.ExternalKafka.Address, cnf.ExternalKafka.User, cnf.ExternalKafka.Pwd, cnf.ExternalKafka.SaslMechanism, runtime.NumCPU(), sarama.V3_6_0_0)

	mysql.InitMysql(cnf)
	data := data2.NewData(mysql.GetDB())

	gin.SetMode(cnf.Http.Mode)
	r := gin.Default()
	r.Use(middleware.Cors())
	r.GET("/health", func(*gin.Context) {})

	api := r.Group("/api")
	api.Use(middleware.Auth())
	cosUploadController := controllers.NewCosUpload(cnf, logger)
	routers.InitCosUploadRouters(api, cosUploadController)

	transformController := controllers.NewTransform(cnf, logger, data)
	routers.InitTransformRouters(api, transformController)

	fs := http.FileServer(http.Dir("www"))
	r.NoRoute(func(ctx *gin.Context) {
		fs.ServeHTTP(ctx.Writer, ctx.Request)
	})
	r.GET("/", func(ctx *gin.Context) {
		http.ServeFile(ctx.Writer, ctx.Request, "www/index.html")
	})

	err := r.Run(fmt.Sprintf("%s:%d", cnf.Http.IP, cnf.Http.Port))
	if err != nil {
		log.Fatal(err)
	}
}
