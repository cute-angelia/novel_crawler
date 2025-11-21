package internal

import (
	"log"
	u "net/url"
	"novel_crawler/consts"
	"novel_crawler/crawler"
	"time"
)

func ConcurrentLimit(urlStr string) {
	glc := make(chan interface{}, 50)
	gap := time.Millisecond * 100

	url, err := u.Parse(urlStr)
	if err != nil {
		log.Fatalln("发生致命错误，请输入正确的链接！！")
	}
	if rf, ok := consts.NewSiteInfoConfigMap[url.Hostname()]; ok {
		glc = make(chan interface{}, rf.Concurrent)
		gap = rf.Gap
		log.Printf("该网站对请求频率进行了限制，本程序的并发量限制为%d， 所以耗时会更长一点", rf.Concurrent)
	}

	*crawler.Glc = glc
	*crawler.Gap = gap
}
