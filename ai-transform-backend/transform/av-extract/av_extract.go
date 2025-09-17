package av_extract

import (
	"ai-transform-backend/message"
	"ai-transform-backend/pkg/config"
	"ai-transform-backend/pkg/constants"
	"ai-transform-backend/pkg/ffmpeg"
	"ai-transform-backend/pkg/log"
	"ai-transform-backend/pkg/mq/kafka"
	"ai-transform-backend/pkg/utils"
	"ai-transform-backend/pkg/zerror"
	_interface "ai-transform-backend/transform/interface"
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"os/exec"
	"path"
	"strings"
	"sync"
)

type avExtract struct {
	conf *config.Config
	log  log.ILogger
}

func NewAvExtract(conf *config.Config, log log.ILogger) _interface.ConsumerTask {
	return &avExtract{
		conf: conf,
		log:  log,
	}
}

func (t *avExtract) Start(ctx context.Context) {
	cg := kafka.NewConsumerGroup(t.conf.ExternalKafka.Address, t.conf.ExternalKafka.User, t.conf.ExternalKafka.Pwd, t.conf.ExternalKafka.SaslMechanism, t.log, sarama.V3_6_0_0, t.messageHandlerFunc)
	cg.Start(ctx, constants.KAFKA_TOPIC_TRANSFORM_AV_EXTRACT, []string{constants.KAFKA_TOPIC_TRANSFORM_AV_EXTRACT})
}

func (t *avExtract) messageHandlerFunc(consumerMessage *sarama.ConsumerMessage) error {
	avExtractMsg := &message.KafkaMsg{}
	err := json.Unmarshal(consumerMessage.Value, avExtractMsg)
	if err != nil {
		t.log.Error(err)
		return err
	}
	t.log.DebugF("%+v \n", avExtractMsg)
	filePath := avExtractMsg.SourceFilePath
	filename := strings.TrimSuffix(path.Base(filePath), path.Ext(filePath))
	videoPath := fmt.Sprintf("%s/%s/%s.mp4", constants.MIDDLE_DIR, filename, filename)
	audioPath := fmt.Sprintf("%s/%s/%s.aac", constants.MIDDLE_DIR, filename, filename)
	err = utils.CreateDirIfNotExists(videoPath, audioPath)
	if err != nil {
		t.log.Error(err)
		return err
	}
	err = t.avExtract(filePath, videoPath, audioPath)
	if err != nil {
		t.log.Error(err)
		return err
	}

	asrMsg := avExtractMsg
	asrMsg.Filename = filename
	asrMsg.ExtractVideoPath = videoPath
	asrMsg.ExtractAudioPath = audioPath

	value, err := json.Marshal(asrMsg)
	if err != nil {
		t.log.Error(err)
		return err
	}
	producerPool := kafka.GetProducerPool(kafka.ProducerPoolKey)
	producer := producerPool.Get()
	defer producerPool.Put(producer)

	msg := &sarama.ProducerMessage{
		Topic: constants.KAFKA_TOPIC_TRANSFORM_ASR,
		Value: sarama.StringEncoder(value),
	}

	_, _, err = producer.SendMessage(msg)
	if err != nil {
		t.log.Error(err)
		return err
	}

	return nil

}

func (t *avExtract) avExtract(filePath, videoPath, audioPath string) (err error) {
	errChan := make(chan error, 2)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		audioCmd := exec.Command(ffmpeg.FFmpeg, "-i", filePath, "-vn", "-acodec", "copy", audioPath)
		t.log.Debug(audioCmd.String())
		err := audioCmd.Run()
		errChan <- err
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		videoCmd := exec.Command(ffmpeg.FFmpeg, "-i", filePath, "-an", "-vcodec", "copy", videoPath)
		t.log.Debug(videoCmd.String())
		err := videoCmd.Run()
		errChan <- err
	}()
	wg.Wait()
	close(errChan)
	err = zerror.NewByErr(<-errChan, <-errChan)
	if err != nil {
		t.log.Error(err)
	}
	return err
}
