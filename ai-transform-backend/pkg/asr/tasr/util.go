package tasr

import (
	"ai-transform-backend/pkg/utils"
	"strconv"
	"strings"
)

func TenCentAsrToSRT(contentStr string) []string {
	contentStr = strings.Trim(strings.Trim(contentStr, " "), "\n")
	contentSlice := strings.Split(contentStr, "\n")
	srtContentSlice := make([]string, 0, 4*len(contentSlice))
	position := 1
	for i := 0; i < len(contentSlice); i++ {
		start, end, content := splitItem(contentSlice[i])
		srtContentSlice = append(srtContentSlice, strconv.Itoa(position), utils.BuildStrItemTimeStr(start, end), content, "")
		position++
	}
	return srtContentSlice
}

func splitItem(str string) (start, end int, content string) {
	timePart := strings.TrimPrefix(strings.Split(str, "] ")[0], "[")
	content = strings.Trim(strings.Replace(str, "["+timePart+"] ", "", 1), " ")
	timeList := strings.Split(timePart, ",")
	startStr := timeList[0]
	endStr := timeList[1]
	start = utils.GetMs(startStr)
	end = utils.GetMs(endStr)
	return
}
