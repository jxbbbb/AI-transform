package constants

// 文件系统目录
const (
	RUNTIME_DIR = "runtime"
	INPUTS_DIR  = RUNTIME_DIR + "/inputs"
	MIDDLE_DIR  = RUNTIME_DIR + "/middle"
	OUTPUTSDIR  = RUNTIME_DIR + "/outputs"
	SRTS_DIR    = RUNTIME_DIR + "/srts"
	REFER_WAV   = RUNTIME_DIR + "/refer"

	AUDIOS_GENERATION_SUB_DIR = "audios"
	TEMP_SUB_DIR              = "tmp"
)

// 对象存储目录
const (
	COS_ROOT      = "/ai-transform"
	COS_INPUT     = COS_ROOT + "/inputs"
	COS_TMP       = COS_ROOT + "/tmp"
	COS_TMP_AUDIO = COS_TMP + "/audios"
	COS_TMP_REFER = COS_TMP + "/refer"
	COS_SRTS      = COS_ROOT + "/srts"
	COS_OUTPUT    = COS_ROOT + "/outputs"
)

// Topic 定义
const (
	// web站入口队列
	KAFKA_TOPIC_TRANSFORM_WEB_ENTRY = "transform_web_entry"

	// 音视频提取
	KAFKA_TOPIC_TRANSFORM_AV_EXTRACT = "transform_av_extract"

	// 字幕识别
	KAFKA_TOPIC_TRANSFORM_ASR = "transform_asr"

	// 字幕翻译
	KAFKA_TOPIC_TRANSFORM_TRANSLATE_SRT = "transform_translate_srt"

	// 根据字幕生成音频
	KAFKA_TOPIC_TRANSFORM_AUDIO_GENERATION = "transform_audio_generation"

	// 音视频合成
	KAFKA_TOPIC_TRANSFORM_AV_SYNTHESIS = "transform_av_synthesis"

	// 保存结果
	KAFKA_TOPIC_TRANSFORM_SAVE_RESULT = "transform_save_result"

	// 参考音频
	KAFKA_TOPIC_TRANSFORM_REFER_WAV = "transform_refer_wav"
)

// 中文
const LANG_ZH = "zh"

// 英文
const LANG_EN = "en"
