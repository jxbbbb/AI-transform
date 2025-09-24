package main

import (
	data2 "ai-transform-backend/data"
	"ai-transform-backend/pkg/asr/tasr"
	"ai-transform-backend/pkg/config"
	"ai-transform-backend/pkg/constants"
	"ai-transform-backend/pkg/db/mysql"
	"ai-transform-backend/pkg/log"
	"ai-transform-backend/pkg/mq/kafka"
	"ai-transform-backend/pkg/storage/cos"
	"ai-transform-backend/pkg/utils"
	"ai-transform-backend/transform/asr"
	av_extract "ai-transform-backend/transform/av-extract"
	"ai-transform-backend/transform/entry"
	refer_wav "ai-transform-backend/transform/refer-wav"
	"context"
	"flag"
	"github.com/IBM/sarama"
	"os"
	"os/signal"
	"runtime"
)

var (
	configFile = flag.String("config", "dev.config.yaml", "")
)

func main() {
	flag.Parse()
	err := utils.CreateDirIfNotExists(
		constants.INPUTS_DIR,
		constants.OUTPUTSDIR,
		constants.MIDDLE_DIR,
		constants.SRTS_DIR,
		constants.REFER_WAV,
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

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
	kafka.InitKafkaProducerPool(kafka.ProducerPoolKey, cnf.Kafka.Address, cnf.Kafka.User, cnf.ExternalKafka.Pwd, cnf.Kafka.SaslMechanism, runtime.NumCPU(), sarama.V3_6_0_0)

	mysql.InitMysql(cnf)
	data := data2.NewData(mysql.GetDB())
	csf := cos.NewCosStorageFactory(
		cnf.Cos.BucketUrl,
		cnf.Cos.SecretId,
		cnf.Cos.SecretKey,
		cnf.Cos.CDNDomain,
	)
	asrfactory := tasr.NewCreateAsrFactory(cnf.Asr.SecretId, cnf.Asr.SecretKey, cnf.Asr.Endpoint, cnf.Asr.Region)
	go entry.NewEntry(cnf, logger, csf).Start(ctx)
	go av_extract.NewAvExtract(cnf, logger).Start(ctx)
	go asr.NewAsr(cnf, logger, csf, data, asrfactory).Start(ctx)
	go refer_wav.NewReferWav(cnf, logger).Start(ctx)
	<-ctx.Done()
}
