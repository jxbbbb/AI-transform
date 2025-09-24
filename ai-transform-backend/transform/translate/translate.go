package translate

import (
	"ai-transform-backend/data"
	"ai-transform-backend/message"
	"ai-transform-backend/pkg/config"
	"ai-transform-backend/pkg/constants"
	"ai-transform-backend/pkg/log"
	machine_translate "ai-transform-backend/pkg/machine-translate"
	"ai-transform-backend/pkg/mq/kafka"
	"ai-transform-backend/pkg/storage"
	"ai-transform-backend/pkg/utils"
	_interface "ai-transform-backend/transform/interface"
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

type translate struct {
	conf              config.Config
	log               log.ILogger
	cosStorageFactory storage.StorageFactory
	data              data.IData
	tf                machine_translate.TranslatorFactory
}

func NewTranslate(conf config.Config, log log.ILogger, cosStorageFactory storage.StorageFactory, tf machine_translate.TranslatorFactory, data data.IData) _interface.ConsumerTask {
	return &translate{
		conf:              conf,
		log:               log,
		cosStorageFactory: cosStorageFactory,
		tf:                tf,
		data:              data,
	}
}

func (t *translate) Start(ctx context.Context) {
	cg := kafka.NewConsumerGroup(t.conf.ExternalKafka.Address, t.conf.ExternalKafka.User, t.conf.ExternalKafka.Pwd, t.conf.ExternalKafka.SaslMechanism, t.log, sarama.V3_6_0_0, t.messageHandlerFunc)
	cg.Start(ctx, constants.KAFKA_TOPIC_TRANSFORM_TRANSLATE_SRT, []string{constants.KAFKA_TOPIC_TRANSFORM_TRANSLATE_SRT})
}

func (t *translate) messageHandlerFunc(consumerMessage *sarama.ConsumerMessage) error {
	translateMsg := &message.KafkaMsg{}
	err := json.Unmarshal(consumerMessage.Value, translateMsg)
	if err != nil {
		t.log.Error(err)
		return err
	}
	t.log.DebugF("%v+\n", translateMsg)
	originalStrPath := translateMsg.OriginalSrtPath
	file, err := os.Open(originalStrPath)
	if err != nil {
		t.log.Error(err)
		return err
	}
	defer file.Close()
	strContentBytes, err := io.ReadAll(file)
	if err != nil {
		t.log.Error(err)
		return err
	}
	strContent := string(strContentBytes)
	srtContentSlice := strings.Split(strContent, "\n")

	err = t.translateSrt(srtContentSlice, translateMsg.SourceLanguage, translateMsg.TargetLanguage)
	if err != nil {
		t.log.Error(err)
		return err
	}
	translateStrFilename := fmt.Sprintf("%s_translate.srt", translateMsg.Filename)
	translateSrtPath := fmt.Sprintf("%s/%s", constants.SRTS_DIR, translateStrFilename)
	err = utils.SaveSrt(srtContentSlice, translateSrtPath)
	if err != nil {
		t.log.Error(err)
		return err
	}
	s := t.cosStorageFactory.CreateStorage()
	storageSrtPath := fmt.Sprintf("%s/%s", constants.COS_SRTS, path.Base(translateSrtPath))
	srtUrl, err := s.UploadFromFile(translateSrtPath, storageSrtPath)
	if err != nil {
		t.log.Error(err)
		return err
	}
	recordsData := t.data.NewTransformRecordsData()
	err = recordsData.Update(&data.TransformRecords{
		ID:               translateMsg.RecordsID,
		TranslatedSrtUrl: srtUrl,
		UpdateAt:         time.Now().Unix(),
	})
	if err != nil {
		t.log.Error(err)
		return err
	}

	generationMsg := translateMsg
	generationMsg.TranslateSrtPath = translateSrtPath
	value, err := json.Marshal(generationMsg)
	if err != nil {
		t.log.Error(err)
		return err
	}

	producerPool := kafka.GetProducerPool(kafka.ProducerPoolKey)
	producer := producerPool.Get()
	defer producerPool.Put(producer)

	msg := &sarama.ProducerMessage{
		Topic: constants.KAFKA_TOPIC_TRANSFORM_AUDIO_GENERATION,
		Value: sarama.StringEncoder(value),
	}

	_, _, err = producer.SendMessage(msg)
	if err != nil {
		t.log.Error(err)
		return err
	}

	return nil
}

func (t *translate) translateSrt(srtContentSlice []string, sourceLanguage, targetLanguage string) error {
	return nil
}
