package controllers

import (
	"ai-transform-backend/data"
	"ai-transform-backend/message"
	"ai-transform-backend/pkg/constants"
	"ai-transform-backend/pkg/mq/kafka"
	"encoding/json"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

type transInfo struct {
	ProjectName       string `form:"project_name" binding:"required" `
	OriginalLanguage  string `form:"original_language" binding:"required"`
	TranslateLanguage string `form:"translate_language" binding:"required"`
	FileUrl           string `form:"file_url" binding:"required,url"`
}

func (c *Transform) Translate(ctx *gin.Context) {
	userID, ok := ctx.Get("User.ID")
	if !ok {
		c.log.Error("鉴权失败")
		ctx.JSON(http.StatusUnauthorized, gin.H{})
		return
	}
	ti := transInfo{}
	err := ctx.ShouldBind(&ti)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entity := &data.TransformRecords{
		UserID:             userID.(int64),
		ProjectName:        ti.ProjectName,
		OriginalLanguage:   ti.OriginalLanguage,
		TranslatedLanguage: ti.TranslateLanguage,
		OriginalVideoUrl:   ti.FileUrl,
		CreateAt:           time.Now().Unix(),
		UpdateAt:           time.Now().Unix(),
	}
	recordData := c.data.NewTransformRecordsData()
	err = recordData.Add(entity)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//将消息推送到kafka
	entryMsg := message.KafkaMsg{
		RecordsID:        entity.ID,
		UserID:           entity.UserID,
		OriginalVideoUrl: ti.FileUrl,
		SourceFilePath:   ti.OriginalLanguage,
	}
	value, err := json.Marshal(entryMsg)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	producePool := kafka.GetProducerPool(kafka.ProducerPoolExternalKey)
	producer := producePool.Get()
	defer producer.Close()
	msg := &sarama.ProducerMessage{
		Topic: constants.KAFKA_TOPIC_TRANSFORM_WEB_ENTRY,
		Value: sarama.StringEncoder(value),
	}

	_, _, err = producer.SendMessage(msg)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "翻译任务已提交",
		"project_id": entity.ID,
	})
}
