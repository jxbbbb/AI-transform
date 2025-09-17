package message

type KafkaMsg struct {
	RecordsID        int64  `json:"recordsID,omitempty"`
	UserID           int64  `json:"userID,omitempty"`
	OriginalVideoUrl string `json:"originalVideoUrl,omitempty"`
	SourceLanguage   string `json:"sourceLanguage,omitempty"`
	TargetLanguage   string `json:"targetLanguage,omitempty"`
	//源文件路径
	SourceFilePath string `json:"sourceFilePath,omitempty"`
	//文件名称，不包含扩展名
	Filename string `json:"filename,omitempty"`
	//提取后的视频文件路径
	ExtractVideoPath string `json:"extractVideoPath,omitempty"`
	//提取后的音频文件路径
	ExtractAudioPath string `json:"extractAudioPath,omitempty"`
	//识别的原始字幕文件路径
	OriginalSrtPath string `json:"originalSrtPath,omitempty"`
	//翻译后的字幕文件路径
	TranslateSrtPath string `json:"translateSrtPath,omitempty"`
	//翻译后字幕裁切后的字幕文件
	TranslateSplitSrtPath string `json:"translateSplitSrtPath,omitempty"`
	//生成的音频文件目录
	GenerationAudioDir string `json:"generationAudioDir,omitempty"`
	//最终输出结果
	OutPutFilePath string `json:"outPutFilePath,omitempty"`
	//参考音频
	ReferWavPath string `json:"referWavPath,omitempty"`
	//参考音频对应的文本
	PromptText string `json:"promptText,omitempty"`
	//参考音频对应的语言
	PromptLanguage string `json:"promptLanguage,omitempty"`
}
