package controllers

import (
	asr2 "ai-transform-backend/pkg/asr"
	"ai-transform-backend/pkg/asr/tasr"
	"ai-transform-backend/pkg/config"
	"ai-transform-backend/pkg/constants"
	"ai-transform-backend/pkg/log"
	"ai-transform-backend/pkg/storage"
	"ai-transform-backend/pkg/zerror"
	"fmt"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

type ReferWav struct {
	conf              *config.Config
	log               log.ILogger
	cosStorageFactory storage.StorageFactory
}

func NewReferWav(cosStorageFactory storage.StorageFactory, conf *config.Config, log log.ILogger) *ReferWav {
	return &ReferWav{
		conf:              conf,
		log:               log,
		cosStorageFactory: cosStorageFactory,
	}
}

type SaveReferInput struct {
	RecordID        int64                 `form:"record_id" binding:"required"`
	PromptText      string                `form:"prompt_text" binding:"required"`
	PromptLanguage  string                `form:"prompt_language" binding:"required"`
	ReferFileHeader *multipart.FileHeader `form:"refer_wav_file" binding:"required"`
}

func (c *ReferWav) SaveReferWav(ctx *gin.Context) {
	in := &SaveReferInput{}
	err := ctx.ShouldBind(in)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	wavFilePath := fmt.Sprintf("%s/%d.wav", constants.REFER_WAV, in.RecordID)
	promptTextFilePath := fmt.Sprintf("%s/%d.txt", constants.REFER_WAV, in.RecordID)

	if c.conf.Http.Mode != gin.ReleaseMode {
		wavFilePath = strings.Replace(wavFilePath, "runtime/refer", "runtime/test-refer", 1)
		promptTextFilePath = strings.Replace(promptTextFilePath, "runtime/refer", "runtime/test-refer", 1)
	}

	err = ctx.SaveUploadedFile(in.ReferFileHeader, wavFilePath)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	//上传音频到cos
	storageAudioPath := fmt.Sprintf("/%s/%s", constants.COS_TMP_REFER, path.Base(wavFilePath))
	s := c.cosStorageFactory.CreateStorage()
	audioUrl, err := s.UploadFromFile(wavFilePath, storageAudioPath)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	//字幕识别
	in.PromptText, err = c.getAsrData(audioUrl)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	file, err := os.OpenFile(promptTextFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer file.Close()
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	content := fmt.Sprintf("%s\n%s", in.PromptLanguage, in.PromptText)
	_, err = file.WriteString(content)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	mp := make(map[string]string, 0)
	mp["refer_wav_path"] = wavFilePath
	mp["prompt_text"] = in.PromptText
	mp["prompt_language"] = in.PromptLanguage
	ctx.JSON(http.StatusOK, mp)
}

func (c *ReferWav) getAsrData(audioUrl string) (string, error) {
	factory := tasr.NewCreateAsrFactory(c.conf.Asr.SecretId, c.conf.Asr.SecretKey, c.conf.Asr.Endpoint, c.conf.Asr.Region)
	a, err := factory.CreateAsr()
	if err != nil {
		return "", zerror.NewByErr(err)
	}
	taskId, err := a.Asr(audioUrl)
	if err != nil {
		return "", zerror.NewByErr(err)
	}
	<-time.After(time.Second * 5)

	result := ""
	status := asr2.FAILED
outloop:
	for {
		result, status, err = a.GetAsrResult(taskId)
		if err != nil {
			return "", zerror.NewByErr(err)
		}
		switch status {
		case asr2.WAITING:
			<-time.After(time.Second * 5)
			continue
		case asr2.DOING:
			<-time.After(time.Second * 1)
			continue
		case asr2.SUCCESS:
			break outloop
		case asr2.FAILED:
			c.log.Error(err)
			return "", zerror.NewByErr(err)
		}
	}
	if result != "" {
		list := strings.Split(result, "\n")
		content := ""
		for i := 0; i < len(list); i++ {
			l := strings.Split(list[i], "]")
			if len(l) > 1 {
				content += strings.Trim(l[1], " ")
			}
		}
		return content, nil
	}
	return "", err
}

type GetReferInput struct {
	RecordID int64 `form:"record_id" binding:"required"`
}

func (c *ReferWav) GetReferInfo(ctx *gin.Context) {
	in := &GetReferInput{}
	err := ctx.ShouldBind(in)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	wavFilePath := fmt.Sprintf("%s/%d.wav", constants.REFER_WAV, in.RecordID)
	promptTextFilePath := fmt.Sprintf("%s/%d.txt", constants.REFER_WAV, in.RecordID)
	if c.conf.Http.Mode != gin.ReleaseMode {
		wavFilePath = strings.Replace(wavFilePath, "runtime/refer", "runtime/test-refer", 1)
		promptTextFilePath = strings.Replace(promptTextFilePath, "runtime/refer", "runtime/test-refer", 1)
	}
	_, err = os.Stat(wavFilePath)
	_, err1 := os.Stat(promptTextFilePath)
	if os.IsNotExist(err) || os.IsNotExist(err1) {
		ctx.JSON(http.StatusOK, gin.H{})
		return
	}
	if err != nil || err1 != nil {
		err = zerror.NewByErr(err, err1)
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	contentBytes, err := os.ReadFile(promptTextFilePath)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	content := string(contentBytes)
	promptLanguage := strings.Split(content, "\n")[0]
	promptText := strings.Split(content, "\n")[1]

	mp := map[string]string{}
	mp["refer_wav_path"] = wavFilePath
	mp["prompt_text"] = promptText
	mp["prompt_language"] = promptLanguage
	ctx.JSON(http.StatusOK, mp)
}
