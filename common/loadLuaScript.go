package common

import (
	"os"
)

// 用绝对路径
func LoadLuaScript(absPath string) (string, error) {
	//读取文件内容
	scriptBytes, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}

	return string(scriptBytes), nil
}
