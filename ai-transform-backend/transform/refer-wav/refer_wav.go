package refer_wav

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
	"bytes"
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/goccy/go-json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// 参考音频时长 ms
const REFER_WAV_DURATION = 6 * 1000

type referWav struct {
	conf *config.Config
	log  log.ILogger
}

func NewReferWav(conf *config.Config, log log.ILogger) _interface.ConsumerTask {
	return &referWav{conf: conf, log: log}
}
func (t *referWav) Start(ctx context.Context) {
	cg := kafka.NewConsumerGroup(
		t.conf.ExternalKafka.Address,
		t.conf.ExternalKafka.User,
		t.conf.ExternalKafka.Pwd,
		t.conf.ExternalKafka.SaslMechanism,
		t.log,
		sarama.V3_6_0_0,
		t.messageHandleFunc,
	)
	cg.Start(ctx, constants.KAFKA_TOPIC_TRANSFORM_REFER_WAV, []string{constants.KAFKA_TOPIC_TRANSFORM_REFER_WAV})
}

func (t *referWav) messageHandleFunc(consumerMessage *sarama.ConsumerMessage) error {
	referMsg := &message.KafkaMsg{}
	err := json.Unmarshal(consumerMessage.Value, referMsg)
	if err != nil {
		t.log.Error(err)
		return err
	}
	t.log.DebugF("%+v\n", referMsg)

	//读取参考音频和文本
	referWavPath, promptText, promptLanguage, err := t.getReferWav(referMsg.RecordsID)
	if err != nil {
		t.log.Error(err)
		return err
	}
	if referWavPath == "" || promptText == "" {
		// 从字幕中获取
		referWavPath, err = t.getReferInfoFromSrt(referMsg.OriginalSrtPath, referMsg.ExtractAudioPath, referMsg.RecordsID)
		if err != nil {
			t.log.Error(err)
			return err
		}
		referWavPath, promptText, promptLanguage, err = t.saveReferWav(referMsg.RecordsID, referWavPath, referMsg.SourceLanguage)
		if err != nil {
			t.log.Error(err)
			return err
		}
	}

	translateMsg := referMsg
	translateMsg.ReferWavPath = referWavPath
	translateMsg.PromptText = promptText
	translateMsg.PromptLanguage = promptLanguage

	value, err := json.Marshal(translateMsg)
	if err != nil {
		t.log.Error(err)
		return err
	}

	producerPool := kafka.GetProducerPool(kafka.ProducerPoolKey)
	producer := producerPool.Get()
	defer producerPool.Put(producer)

	msg := &sarama.ProducerMessage{
		Topic: constants.KAFKA_TOPIC_TRANSFORM_TRANSLATE_SRT,
		Value: sarama.StringEncoder(value),
	}
	_, _, err = producer.SendMessage(msg)
	if err != nil {
		t.log.Error(err)
		return err
	}
	return nil
}

func (t *referWav) getReferInfoFromSrt(originalSrtPath, extractAudioPath string, recordID int64) (referWavPath string, err error) {
	file, err := os.Open(originalSrtPath)
	if err != nil {
		t.log.Error(err)
		return "", err
	}
	defer file.Close()
	srtContentBytes, err := io.ReadAll(file)
	if err != nil {
		t.log.Error(err)
		return "", err
	}
	if len(srtContentBytes) == 0 {
		return "", zerror.NewByMsg("字幕文件读取失败")
	}
	srtContent := string(srtContentBytes)
	srtContentSlice := strings.Split(srtContent, "\n")
	timeSrt := ""
	timeDuration := 0
	for i := 0; i < len(srtContentSlice); i += 4 {
		start, end := utils.GetSrtTime(srtContentSlice[i+1])
		duration := end - start
		if duration >= REFER_WAV_DURATION {
			timeSrt = srtContentSlice[i+1]
			timeDuration = duration
			break
		} else if duration > timeDuration {
			timeSrt = srtContentSlice[i+1]
			timeDuration = duration
		}
	}
	ss := strings.Replace(strings.Split(timeSrt, " --> ")[0], ",", ".", 1)
	d := float64(timeDuration) / float64(1000)
	if timeDuration > REFER_WAV_DURATION {
		d = float64(REFER_WAV_DURATION) / float64(1000)
	}
	referWavPath = fmt.Sprintf("%s/%d.wav", constants.REFER_WAV, recordID)
	err = t.cutReferWav(extractAudioPath, ss, d, referWavPath)
	if err != nil {
		t.log.Error(err)
		return "", err
	}
	return referWavPath, nil
}
func (t *referWav) cutReferWav(originalAudioPath, ss string, d float64, dstAudioPath string) error {
	cmd := exec.Command(ffmpeg.FFmpeg, "-i", originalAudioPath, "-ss", ss, "-t", fmt.Sprintf("%.3f", d), dstAudioPath)
	t.log.Debug(cmd.String())
	i := 0
	var err error
retry:
	err = cmd.Run()
	if err != nil && i < 3 {
		i++
		<-time.After(time.Millisecond * 500)
		goto retry
	}
	return err
}

var client = &http.Client{}

func (t *referWav) getReferWav(recordID int64) (referWavPath, promptText, promptLanguage string, err error) {
	addr := t.conf.DependOn.ReferWav.Address
	url := fmt.Sprintf("%s/api/refer/wav?record_id=%d", addr, recordID)
	method := "GET"
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.log.Error(err)
		return
	}
	res, err := client.Do(req)
	if err != nil {
		t.log.Error(err)
		return
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.log.Error(err)
		return
	}
	mp := make(map[string]string)
	err = json.Unmarshal(body, &mp)
	if err != nil {
		t.log.Error(err)
		return
	}
	return mp["refer_wav_path"], mp["prompt_text"], mp["prompt_language"], nil
}
func (t *referWav) saveReferWav(recordID int64, referWavPathIn, promptLanguageIn string) (referWavPath, promptText, promptLanguage string, err error) {
	addr := t.conf.DependOn.ReferWav.Address
	url := fmt.Sprintf("%s/api/refer/wav", addr)
	method := "POST"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("record_id", fmt.Sprintf("%d", recordID))
	_ = writer.WriteField("prompt_text", "参考音频文本，已弃用")
	_ = writer.WriteField("prompt_language", promptLanguageIn)
	file, err := os.Open(referWavPathIn)
	if err != nil {
		t.log.Error(err)
		return
	}
	defer file.Close()
	part3, err := writer.CreateFormFile("refer_wav_file", filepath.Base(referWavPathIn))
	if err != nil {
		t.log.Error(err)
		return
	}
	_, err = io.Copy(part3, file)
	if err != nil {
		t.log.Error(err)
		return
	}

	err = writer.Close()
	if err != nil {
		t.log.Error(err)
		return
	}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		t.log.Error(err)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		t.log.Error(err)
		return
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.log.Error(err)
		return
	}
	mp := make(map[string]string)
	err = json.Unmarshal(body, &mp)
	if err != nil {
		t.log.Error(err)
		return
	}
	return mp["refer_wav_path"], mp["prompt_text"], mp["prompt_language"], nil
}
