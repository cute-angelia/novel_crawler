package utils

import (
	"fmt"
	"github.com/spf13/viper"
	"strconv"
	"strings"
)

const (
	textBlack = iota + 30
	textRed
	textGreen
	textYellow
	textBlue
	textPurple
	textCyan
	textWhite
)

func Purple(str string) string {
	return textColor(textPurple, str)
}
func Yellow(str string) string {
	return textColor(textYellow, str)
}
func Red(str string) string {
	return textColor(textRed, str)
}

func Green(str string) string {
	return textColor(textGreen, str)
}

func textColor(color int, str string) string {
	return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", color, str)
}

func IsDebug() bool {
	return viper.GetBool("common.debug")
}

// str := "0-0"
func RangeConvert(str string) []int {
	// 使用 strings.Split 分割字符串
	parts := strings.Split(str, "-")
	// 确保分割后有正确的部分
	if len(parts) == 2 {
		// 将分割后的每个部分转换为整数
		num1, err1 := strconv.Atoi(parts[0])
		num2, err2 := strconv.Atoi(parts[1])

		// 检查转换是否成功，并创建整数切片
		if err1 == nil && err2 == nil {
			result := []int{num1, num2}
			return result
		} else {
			return []int{0, 0}
		}
	} else {
		return []int{0, 0}
	}
}
