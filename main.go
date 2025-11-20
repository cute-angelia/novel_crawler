package main

import (
	"flag"
	"github.com/cute-angelia/go-utils/components/loggers/loggerV3"
	"github.com/cute-angelia/go-utils/utils/conf"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"log"
	u "net/url"
	"novel_crawler/consts"
	"novel_crawler/crawler"
	"novel_crawler/utils"
	"os"
	"sort"
	"sync"
	"time"
)

// 目前适配网站 https://www.52bqg.org/book_128955/

func retry(task func() error, count int) error {
	if err := task(); err == nil {
		return nil
	} else if count > 1 {
		//log.Println("do retry, remain retry count:", count-1)
		return retry(task, count-1)
	} else {
		return err
	}

}

func initConcurrentLimit(urlStr string) {
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

// doCrawler 控制爬取流程
func doCrawler(urlStr, fileName string) {
	if c, err := crawler.CreateCrawler(urlStr); err == nil {

		log.Println("正在获取章节列表......")

		if chapters, err := c.FetchChapterList(); err == nil {
			log.Println(utils.Green("章节列表已获取"))
			log.Println("正在下载章节内容......")

			// 创建文件
			file, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					log.Println(utils.Red("Error: " + err.Error() + "\n"))
				}
			}(file)

			// 这里也要限制一下并发量，为什么呢，因为有些章节是分页展示的，如果过这里不限制并发量，所有章节的所有页面都随机地获取
			// 容易出现爬取的页面虽然很多，但爬取的完整章节很少的情况。这时候在前期进度条就会始终显示为0，虽然爬取总时间不变，用户体验感不好。
			glc := make(chan interface{}, 50)
			if rf, ok := consts.NewSiteInfoConfigMap[c.GetUrl().Hostname()]; ok {
				glc = make(chan interface{}, rf.Concurrent)
				chapters = crawler.ExtractRange(chapters, rf.ChapterListRange)

				if utils.IsDebug() {
					log.Println(chapters)
				}
			}

			// 进度条，进度条每次输出时，会把上一行消除掉，所以打日志时每行末尾多加一个\n
			p := mpb.New(mpb.WithWidth(64))
			bar := p.New(int64(len(chapters)),
				// BarFillerBuilder with custom style
				mpb.BarStyle().Lbound("╢").Filler("=").Tip(">").Padding("-").Rbound("╟"),
				mpb.PrependDecorators(
					decor.Name(utils.Green("章节下载中......"), decor.WC{W: len("章节下载中......") + 1, C: decor.DidentRight}),
					decor.Name(utils.Green("进度："), decor.WCSyncSpaceR),
					decor.CountersNoUnit(utils.Green("%d / %d"), decor.WCSyncWidth),
				),
				mpb.AppendDecorators(
					decor.OnComplete(decor.Percentage(decor.WC{W: 5}), "done"),
				),
			)

			// 爬取每一章节的内容
			// 使用 sync.WaitGroup 确保所有协程完成
			var wg sync.WaitGroup
			errChapters := make([]*crawler.Chapter, 0)
			var mu sync.Mutex // 保护 errChapters 的并发访问

			// 使用带缓冲的通道（高性能）
			// 对于大量章节，可以使用通道来收集结果
			// 使用通道收集成功章节
			successChan := make(chan struct {
				idx     int
				chapter *crawler.Chapter
			}, len(chapters))

			for i := 0; i < len(chapters); i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()

					glc <- 1
					defer func() { <-glc }()

					err = retry(func() error {
						return c.FetchChapterContent(&chapters[idx])
					}, 5)

					if err != nil {
						log.Println(utils.Red("Error: " + err.Error() + "\n"))
						mu.Lock()
						errChapters = append(errChapters, &chapters[idx])
						mu.Unlock()
					} else {
						// 成功抓取后立即写入文件
						// 发送到通道
						successChan <- struct {
							idx     int
							chapter *crawler.Chapter
						}{idx, &chapters[idx]}
					}
					// 等待所有下载完成
					bar.Increment()                    // 确保在 defer 中调用
					time.Sleep(time.Millisecond * 100) // 休眠0.1秒，让控制台io同步
				}(i)
			}

			// 启动一个专门的goroutine来处理写入
			go func() {
				wg.Wait()
				close(successChan)
			}()

			// 收集所有成功章节并按顺序写入
			successList := make([]struct {
				idx     int
				chapter *crawler.Chapter
			}, 0, len(chapters))

			for item := range successChan {
				successList = append(successList, item)
			}

			// 按索引排序并写入
			sort.Slice(successList, func(i, j int) bool {
				return successList[i].idx < successList[j].idx
			})

			for _, item := range successList {
				err = item.chapter.Save(file)
				if err != nil {
					log.Println(utils.Red("写入错误: " + err.Error() + "\n"))
				}
			}
			p.Wait()

			// 提示错误
			if len(errChapters) > 0 {
				log.Println(utils.Red("由于某些原因，以下章节爬取过程出现错误："))
				for _, ec := range errChapters {
					log.Println(utils.Red(ec.Title))
				}
			}

			log.Println(utils.Green("所有章节爬取完毕......"))
			log.Println("正在把爬取结果写入文件......")
			//for _, cha := range chapters {
			//	err = cha.Save(file)
			//	if err != nil {
			//		log.Println(utils.Red("Error: " + err.Error() + "\n"))
			//	}
			//}
			log.Println(utils.Green("程序已运行结束"))
		} else {
			log.Println(utils.Red("Error x: " + err.Error() + "\n"))
		}

	} else {
		log.Println(utils.Red("Error: " + err.Error() + "\n"))
	}

}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println(utils.Yellow("注意，如果程序超过一分钟无响应，请重新执行"))
	var fileName = flag.String("f", "", "保存文件名")
	var urlStr = flag.String("u", "", "url链接")
	flag.Parse()

	loggerV3.New(loggerV3.WithIsOnline(false))

	// 日志
	conf.MergeConfigWithPath("./")

	initConcurrentLimit(*urlStr)
	doCrawler(*urlStr, *fileName+".txt")

	time.Sleep(time.Second)
}
