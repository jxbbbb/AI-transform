package asr

type Asr interface {
	Asr(url string) (taskID uint64, err error)
	GetAsrResult(taskID uint64) (result string, status AsrStatus, err error)
}
type AsrFactory interface {
	CreateAsr() (Asr, error)
}

type AsrStatus string

const (
	SUCCESS AsrStatus = "success"
	WAITING AsrStatus = "waiting"
	DOING   AsrStatus = "doing"
	FAILED  AsrStatus = "failed"
)
