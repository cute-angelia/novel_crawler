package main

import (
	"flag"
	"fmt"
	"github.com/cute-angelia/go-utils/components/loggers/loggerV3"
	"github.com/cute-angelia/go-utils/third_party/weworkrobot"
	"github.com/cute-angelia/go-utils/utils/conf"
	"github.com/spf13/viper"
	"log"
	"novel_crawler/internal"
	"novel_crawler/pkg/utils"
	"time"
)

// 目前适配网站 https://www.52bqg.org/book_128955/

func main() {
	log.Println(utils.Yellow("注意，如果程序超过一分钟无响应，请重新执行"))
	var fileName = flag.String("f", "", "保存文件名")
	var urlStr = flag.String("u", "", "url链接")
	var nums = flag.String("n", "0-0", "下载多少章")
	flag.Parse()

	loggerV3.New(loggerV3.WithIsOnline(false))

	// 日志
	conf.MergeConfigWithPath("./")

	// 限流
	internal.ConcurrentLimit(*urlStr)

	// 抓取
	bookName, total := internal.DoCrawler(*urlStr, *fileName, *nums)

	time.Sleep(time.Second)

	// notify
	weworkrobot.Load(viper.GetString("common.robotKey")).Build(weworkrobot.WithTopic("下载书籍完成")).SendText(fmt.Sprintf("\n下载书籍完成\n书名：《%s》 \n章节：%d章", bookName, total))
}
