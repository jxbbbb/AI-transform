package kafka

import (
	"ai-transform-backend/pkg/log"
	"context"
	"errors"
	"github.com/IBM/sarama"
)

type ConsumerGroup interface {
	Start(ctx context.Context, groupID string, topics []string)
}
type MessageHandleFunc func(message *sarama.ConsumerMessage) error
type consumerGroup struct {
	log               log.ILogger
	brokerList        []string
	user              string
	pwd               string
	saslMechanism     string
	version           sarama.KafkaVersion
	messageHandleFunc MessageHandleFunc
}

func NewConsumerGroup(brokerList []string, user, pwd, saslMechanism string, log log.ILogger, version sarama.KafkaVersion, messageHandleFunc MessageHandleFunc) ConsumerGroup {
	return &consumerGroup{
		log:               log,
		brokerList:        brokerList,
		user:              user,
		pwd:               pwd,
		saslMechanism:     saslMechanism,
		version:           version,
		messageHandleFunc: messageHandleFunc,
	}
}

func (cg *consumerGroup) Start(ctx context.Context, groupID string, topics []string) {
	config := sarama.NewConfig()
	config.Version = cg.version
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Net.SASL.Enable = true
	config.Net.SASL.User = cg.user
	config.Net.SASL.Password = cg.pwd
	config.Net.SASL.Mechanism = sarama.SASLMechanism(cg.saslMechanism)

	client, err := sarama.NewConsumerGroup(cg.brokerList, groupID, config)
	if err != nil {
		log.Error(err)
		return
	}
	cgh := &consumerGroupHandler{
		messageHandleFunc: cg.messageHandleFunc,
	}
	for {
		if err := client.Consume(ctx, topics, cgh); err != nil {
			if errors.Is(err, sarama.ErrClosedConsumerGroup) {
				return
			}
			log.Error(err)
		}
		if ctx.Err() != nil {
			return
		}
	}
}

type consumerGroupHandler struct {
	messageHandleFunc MessageHandleFunc
}

func (cgh *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}
func (cgh *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}
func (cgh *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			{
				if !ok {
					log.Info("消息通道已关闭")
					return nil
				}
				err := cgh.messageHandleFunc(message)
				if err != nil {
					log.Error(err)
				}
				session.MarkMessage(message, "")
			}
		case <-session.Context().Done():
			{
				return nil
			}
		}
	}
}
