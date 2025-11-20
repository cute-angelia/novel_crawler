package utils

import (
	"fmt"
	"github.com/spf13/viper"
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

// ExtractRange 从切片中提取指定范围的元素，自动处理超出边界情况
func ExtractRange(ilist []int, rangeSpec []int) []int {
	if len(rangeSpec) != 2 {
		return []int{}
	}

	start := rangeSpec[0]
	end := rangeSpec[1]

	// 处理起始位置超出边界的情况
	if start < 0 {
		start = 0
	}
	if start > len(ilist) {
		start = len(ilist)
	}

	// 处理结束位置超出边界的情况
	if end < start {
		end = start
	}
	if end > len(ilist) {
		end = len(ilist)
	}

	// 使用copy函数确保顺序一致且避免共享底层数组
	result := make([]int, end-start)
	copy(result, ilist[start:end])

	return result
}

func IsDebug() bool {
	return viper.GetBool("common.debug")
}
