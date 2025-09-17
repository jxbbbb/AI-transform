package kafka

import (
	"ai-transform-backend/pkg/log"
	"github.com/IBM/sarama"
	"sync"
)

const ProducerPoolKey = "kafka"
const ProducerPoolExternalKey = "external_kafka"

var _producerPoolMap map[string]ProducerPool

type ProducerPool interface {
	Get() sarama.SyncProducer
	Put(sarama.SyncProducer)
}
type producerPool struct {
	mutex     sync.Mutex
	producers []sarama.SyncProducer
	maxNum    int
	currIndex int

	brokerList               []string
	user, pwd, saslMechanism string
	version                  sarama.KafkaVersion
}

func InitKafkaProducerPool(key string, brokerList []string, user, pwd, saslMechanism string, maxNum int, version sarama.KafkaVersion) {
	if _producerPoolMap == nil {
		_producerPoolMap = make(map[string]ProducerPool, 0)
	}
	_producerPoolMap[key] = newProducerPool(brokerList, user, pwd, saslMechanism, maxNum, version)
}

func newProducerPool(brokerList []string, user, pwd, saslMechanism string, maxNum int, version sarama.KafkaVersion) ProducerPool {
	if maxNum <= 0 {
		maxNum = 1
	}
	return &producerPool{
		mutex:         sync.Mutex{},
		producers:     make([]sarama.SyncProducer, maxNum),
		maxNum:        maxNum,
		currIndex:     0,
		brokerList:    brokerList,
		user:          user,
		pwd:           pwd,
		saslMechanism: saslMechanism,
		version:       version,
	}
}

func (pp *producerPool) Get() sarama.SyncProducer {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()

	pp.currIndex += 1
	if pp.currIndex >= pp.maxNum {
		pp.currIndex = 0
	}
	producer := pp.producers[pp.currIndex]

	if producer == nil {
		producer = pp.new()
		pp.producers[pp.currIndex] = producer
	}
	return producer
}
func (pp *producerPool) Put(producer sarama.SyncProducer) {}

func (pp *producerPool) new() sarama.SyncProducer {
	config := sarama.NewConfig()
	config.Version = pp.version
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true
	config.Net.SASL.Enable = true
	config.Net.SASL.User = pp.user
	config.Net.SASL.Password = pp.pwd
	config.Net.SASL.Mechanism = sarama.SASLMechanism(pp.saslMechanism)
	producer, err := sarama.NewSyncProducer(pp.brokerList, config)
	if err != nil {
		log.Error(err)
		return nil
	}
	return producer
}

func GetProducerPool(key string) ProducerPool {
	return _producerPoolMap[key]
}
