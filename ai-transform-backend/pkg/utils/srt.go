package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func SaveSrt(srtContentSlice []string, outputFile string) error {
	srtContent := strings.Join(srtContentSlice, "\n")
	file, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(srtContent)
	if err != nil {
		return err
	}
	return nil
}
func GetSrtTime(timeStr string) (start, end int) {
	// 00:29:20,720 --> 00:29:23,560
	l := strings.Split(timeStr, " --> ")
	return GetMs(l[0]), GetMs(l[1])
}

// 获取时间戳的毫秒数
// 00:29:20.720 或者 00:29:20,720
func GetMs(timeStr string) int {
	timeStr = strings.Replace(timeStr, ",", ".", 1)
	l := strings.Split(timeStr, ":")
	if len(l) == 2 {
		l = append([]string{"00"}, l...)
	}
	h, _ := strconv.Atoi(l[0])
	m, _ := strconv.Atoi(l[1])
	s, _ := strconv.ParseFloat(l[2], 64)
	ms := h*3600*1000 + m*60*1000 + int(s*1000)
	return ms
}

func BuildStrItemTimeStr(start, end int) string {
	startStr := toStrTimeStr(start)
	endStr := toStrTimeStr(end)
	return fmt.Sprintf("%s --> %s", startStr, endStr)
}

func toStrTimeStr(ms int) string {
	h := ms / 3600 / 1000

	ms = ms % (3600 * 1000)
	m := ms / 60 / 1000

	ms = ms % (60 * 1000)
	s := ms / 1000

	ms = ms % 1000

	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}
