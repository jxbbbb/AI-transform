package utils

import (
	"os"
	"path"
)

func CreateDirIfNotExists(dirPath ...string) error {
	for _, p := range dirPath {
		dir := p
		ext := path.Ext(p)
		if ext != "" {
			dir = path.Dir(p)
		}
		// 检查文件夹是否存在
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0644)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
