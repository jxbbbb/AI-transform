package audio_generation

import (
	"ai-transform-backend/message"
	"ai-transform-backend/pkg/config"
	"ai-transform-backend/pkg/constants"
	go_pool "ai-transform-backend/pkg/go-pool"
	"ai-transform-backend/pkg/log"
	"ai-transform-backend/pkg/mq/kafka"
	"ai-transform-backend/pkg/utils"
	"ai-transform-backend/pkg/zerror"
	_interface "ai-transform-backend/transform/interface"
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
)

const audioTs = 5 * 1000

type generation struct {
	conf *config.Config
	log  log.ILogger
}

func NewGeneration(conf *config.Config, log log.ILogger) _interface.ConsumerTask {
	return &generation{
		conf: conf,
		log:  log,
	}
}

func (t *generation) Start(ctx context.Context) {
	cg := kafka.NewConsumerGroup(
		t.conf.ExternalKafka.Address,
		t.conf.ExternalKafka.User,
		t.conf.ExternalKafka.Pwd,
		t.conf.ExternalKafka.SaslMechanism,
		t.log,
		sarama.V3_6_0_0,
		t.messageHandleFunc,
	)
	cg.Start(ctx, constants.KAFKA_TOPIC_TRANSFORM_AUDIO_GENERATION, []string{constants.KAFKA_TOPIC_TRANSFORM_AUDIO_GENERATION})
}

func (t *generation) messageHandleFunc(consumerMessage *sarama.ConsumerMessage) error {
	generationMsg := &message.KafkaMsg{}
	err := json.Unmarshal(consumerMessage.Value, generationMsg)
	if err != nil {
		t.log.Error(err)
		return err
	}
	t.log.DebugF("%+v \n", generationMsg)
	translateSrtPath := generationMsg.TranslateSrtPath
	file, err := os.Open(translateSrtPath)
	if err != nil {
		t.log.Error(err)
		return err
	}
	defer file.Close()

	srtContentBytes, err := io.ReadAll(file)
	if err != nil {
		t.log.Error(err)
		return err
	}
	srtContentSlice := strings.Split(string(srtContentBytes), "\n")
	srtContentSlice = splitSrtContent(srtContentSlice, generationMsg.TargetLanguage)

	translateSplitSrtFilename := fmt.Sprintf("%s_split.srt", generationMsg.Filename)
	translateSplitSrtPath := fmt.Sprintf("%s/%s", constants.SRTS_DIR, translateSplitSrtFilename)
	err = utils.SaveSrt(srtContentSlice, translateSplitSrtPath)
	if err != nil {
		t.log.Error(err)
		return err
	}

	dstDir := fmt.Sprintf("%s/%s/%s", constants.MIDDLE_DIR, generationMsg.Filename, constants.AUDIOS_GENERATION_SUB_DIR)
	err = utils.CreateDirIfNotExists(dstDir)
	if err != nil {
		t.log.Error(err)
		return err
	}

	err = t.generateAudio(srtContentSlice, dstDir, "wav", generationMsg.TargetLanguage, generationMsg.ReferWavPath, generationMsg.PromptText, generationMsg.PromptLanguage)
	if err != nil {
		t.log.Error(err)
		return err
	}

	avSynthesisMsg := generationMsg
	avSynthesisMsg.GenerationAudioDir = dstDir
	avSynthesisMsg.TranslateSplitSrtPath = translateSplitSrtPath
	value, err := json.Marshal(avSynthesisMsg)
	if err != nil {
		t.log.Error(err)
		return err
	}
	producerPool := kafka.GetProducerPool(kafka.ProducerPoolKey)
	producer := producerPool.Get()
	defer producerPool.Put(producer)

	msg := &sarama.ProducerMessage{
		Topic: constants.KAFKA_TOPIC_TRANSFORM_AV_SYNTHESIS,
		Value: sarama.StringEncoder(value),
	}
	_, _, err = producer.SendMessage(msg)
	if err != nil {
		t.log.Error(err)
		return err
	}
	return nil

}
func (t *generation) generateAudio(srtContentSlice []string, outputPath, format string, textLanguage, referWavPath, promptText, promptLanguage string) error {
	errChan := make(chan error, len(srtContentSlice)/4)
	executors := make([]go_pool.IExecutor, len(t.conf.DependOn.GPT))
	for i := 0; i < len(t.conf.DependOn.GPT); i++ {
		reqUrl := t.conf.DependOn.GPT[i]
		executors[i] = NewAudioReasoningExecutor(reqUrl, errChan)
	}
	pool := go_pool.NewPool(len(executors), executors...)
	pool.Start()
	for i := 0; i < len(srtContentSlice); i += 4 {
		output := fmt.Sprintf("%s/%s.%s", outputPath, srtContentSlice[i], format)
		text := srtContentSlice[i+2]
		params := &audioReasoningParams{
			text:           text,
			textLanguage:   textLanguage,
			output:         output,
			referWavPath:   referWavPath,
			promptLanguage: promptLanguage,
			promptText:     promptText,
		}
		t := newTask(params)
		pool.Schedule(t)
	}
	pool.WaitAndClose()
	close(errChan)
	errs := make([]error, 0)
	for err := range errChan {
		if err != nil {
			t.log.Error(err)
			errs = append(errs, err)
		}
	}
	err := zerror.NewByErr(errs...)
	return err
}

type audioReasoningExecutor struct {
	reqUrl     string
	httpClient *http.Client
	errChan    chan error
}

func NewAudioReasoningExecutor(reqUrl string, errChan chan error) go_pool.IExecutor {
	return &audioReasoningExecutor{
		reqUrl:     reqUrl,
		httpClient: &http.Client{},
		errChan:    errChan,
	}
}

func (e *audioReasoningExecutor) Exec(t go_pool.ITask) {
	params := t.Run().(*audioReasoningParams)
	j := 0
retry:
	err := audioReasoning(e.httpClient, e.reqUrl, params.text, params.textLanguage, params.output, params.referWavPath, params.promptText, params.promptLanguage)
	if err != nil {
		if j < 3 {
			j++
			goto retry
		}
		e.errChan <- err
	}
}

type audioReasoningParams struct {
	text           string
	textLanguage   string
	output         string
	referWavPath   string
	promptText     string
	promptLanguage string
}
type task struct {
	params *audioReasoningParams
}

func newTask(params *audioReasoningParams) go_pool.ITask {
	return &task{
		params: params,
	}
}

func (t *task) Run() any {
	return t.params
}

func splitSrtContent(srtContentSlice []string, lang string) []string {
	position := 1
	list := make([]string, 0, len(srtContentSlice))
	for i := 0; i < len(srtContentSlice); i += 4 {
		timeStr := srtContentSlice[i+1]
		content := srtContentSlice[i+2]
		start, end := utils.GetSrtTime(timeStr)
		duration := end - start
		if duration > audioTs {
			l, p := splitItemContent(start, end, position, content, lang)
			list = append(list, l...)
			position = p
			continue
		}
		list = append(list, fmt.Sprintf("%d", position), timeStr, content, "")
		position++
	}
	return list
}

func splitItemContent(start, end, position int, content, lang string) ([]string, int) {
	result := make([]string, 0)
	count := int(math.Round(float64(end-start) / float64(audioTs)))
	if count == 0 || count == 1 {
		timeStr := utils.BuildStrItemTimeStr(start, end)
		result = append(result, fmt.Sprintf("%d", position), timeStr, content, "")
		position++
		return result, position
	}
	list := convertStringSlice(content, lang)

	totalLen := len(list)
	mode := totalLen % count
	itemLen := totalLen / count
	itemTS := (end - start) / count
	tsMode := (end - start) % count

	nextIndex := 0
	nextStart := start
	for nextIndex < len(list) {
		endIndex := nextIndex + itemLen
		if mode != 0 {
			endIndex += 1
			mode--
		}
		l := list[nextIndex:endIndex]
		c := convertSliceString(l, lang)
		nextIndex = endIndex

		s := nextStart
		e := s + itemTS
		if tsMode != 0 {
			e += 1
			tsMode--
		}
		if e > end {
			log.WarningF("截止时间计算有误，预期为：%d，实际为：%d", end, e)
			e = end
		}
		nextStart = e
		result = append(result, fmt.Sprintf("%d", position), utils.BuildStrItemTimeStr(s, e), c, "")
		position++
	}
	return result, position
}
func convertStringSlice(content, lang string) []string {
	switch lang {
	case constants.LANG_ZH:
		runes := []rune(content)
		list := make([]string, len(runes))
		for i, r := range runes {
			list[i] = string(r)
		}
		return list
	case constants.LANG_EN:
		return strings.Fields(content)
	default:
		return strings.Fields(content)
	}
}

func convertSliceString(list []string, lang string) string {
	switch lang {
	case constants.LANG_ZH:
		return strings.Join(list, "")
	case constants.LANG_EN:
		return strings.Join(list, " ")
	default:
		return strings.Join(list, " ")
	}
}

func audioReasoning(httpClient *http.Client, reqUrl, text, textLanguage, output, referWavPath, promptText, promptLanguage string) error {
	url := reqUrl
	method := "POST"
	mp := map[string]string{
		"refer_wav_path":  referWavPath,
		"prompt_text":     promptText,
		"prompt_language": promptLanguage,
		"text":            text,
		"text_language":   textLanguage,
	}
	bytes, _ := json.Marshal(mp)
	payload := strings.NewReader(string(bytes))
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(body)
	if err != nil {
		return err
	}
	return nil
}
