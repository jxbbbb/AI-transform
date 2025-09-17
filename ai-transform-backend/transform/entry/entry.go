package entry

import (
	"ai-transform-backend/message"
	"ai-transform-backend/pkg/config"
	"ai-transform-backend/pkg/constants"
	"ai-transform-backend/pkg/log"
	"ai-transform-backend/pkg/mq/kafka"
	"ai-transform-backend/pkg/storage"
	_interface "ai-transform-backend/transform/interface"
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"net/url"
	"path"
	"strings"
)

type entry struct {
	conf              *config.Config
	log               log.ILogger
	cosStorageFactory storage.StorageFactory
}

func NewEntry(conf *config.Config, log log.ILogger, cosStorageFactory storage.StorageFactory) _interface.ConsumerTask {
	return &entry{
		conf:              conf,
		log:               log,
		cosStorageFactory: cosStorageFactory,
	}
}

func (t *entry) Start(ctx context.Context) {
	cg := kafka.NewConsumerGroup(t.conf.ExternalKafka.Address, t.conf.ExternalKafka.User, t.conf.ExternalKafka.Pwd, t.conf.ExternalKafka.SaslMechanism, t.log, sarama.V3_6_0_0, t.messageHandlerFunc)
	cg.Start(ctx, constants.KAFKA_TOPIC_TRANSFORM_WEB_ENTRY, []string{constants.KAFKA_TOPIC_TRANSFORM_WEB_ENTRY})
}

func (t *entry) messageHandlerFunc(consumerMessage *sarama.ConsumerMessage) error {
	entryMsg := &message.KafkaMsg{}
	err := json.Unmarshal(consumerMessage.Value, entryMsg)
	if err != nil {
		t.log.Error(err)
		return err
	}
	t.log.DebugF("%+v \n", entryMsg)
	cs := t.cosStorageFactory.CreateStorage()
	dstPath := fmt.Sprintf("%s/%s", constants.INPUTS_DIR, path.Base(entryMsg.OriginalVideoUrl))
	originalVideoUrl, err := url.Parse(entryMsg.OriginalVideoUrl)
	if err != nil {
		t.log.Error(err)
		return err
	}
	objectKey := strings.Trim(originalVideoUrl.Path, "/")
	err = cs.DownloadFile(objectKey, dstPath)
	if err != nil {
		t.log.Error(err)
		return err
	}
	avExtractMsg := entryMsg
	avExtractMsg.SourceFilePath = dstPath
	value, err := json.Marshal(avExtractMsg)
	if err != nil {
		t.log.Error(err)
		return err
	}
	producerPool := kafka.GetProducerPool(kafka.ProducerPoolKey)
	producer := producerPool.Get()
	defer producerPool.Put(producer)

	msg := &sarama.ProducerMessage{
		Topic: constants.KAFKA_TOPIC_TRANSFORM_AV_EXTRACT,
		Value: sarama.StringEncoder(value),
	}

	_, _, err = producer.SendMessage(msg)
	if err != nil {
		t.log.Error(err)
		return err
	}

	return nil

}
