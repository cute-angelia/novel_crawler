package crawler

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/guonaihong/gout"
	"github.com/spf13/viper"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"log"
	"net/http"
	u "net/url"
	"novel_crawler/consts"
	"novel_crawler/pkg/goutclient"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

// Chapter 爬取流程相同的共用同一个实现类
type Chapter struct {
	Number                int
	UrlStr                string
	Title                 string
	Content               string
	ChapterTitleInContent bool
}

func (c *Chapter) Save(f *os.File) error {
	if c.ChapterTitleInContent {
		_, err := f.WriteString(fmt.Sprintf("%s\n", c.Content))
		return err
	} else {
		_, err := f.WriteString(fmt.Sprintf("%s\n%s\n", c.Title, c.Content))
		return err
	}
}

// ExtractRange 从切片中提取指定范围的元素，自动处理超出边界情况
func ExtractRange(ilist []Chapter, rangeSpec []int) []Chapter {
	if len(rangeSpec) != 2 {
		return []Chapter{}
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
	if end > len(ilist) || end == 0 || end == -1 {
		end = len(ilist)
	}

	// 使用copy函数确保顺序一致且避免共享底层数组
	result := make([]Chapter, end-start)
	copy(result, ilist[start:end])

	return result
}

type ChapterFilter interface {
	Filter(chapters []Chapter) []Chapter
}

type NextGetter interface {
	NextUrl(dom *goquery.Document, selector, subStr string) (*u.URL, error)
}

type CrawlerInterface interface {
	// FetchChapterList 获取章节列表
	FetchChapterList() ([]Chapter, error)
	// FetchChapterContent 获取某一章节内容
	FetchChapterContent(c *Chapter) error
	// GetUrl 获取url
	GetUrl() *u.URL
}

var client = &http.Client{
	Timeout: time.Second * 5,
}

// Glc goroutine limit channel 限制并发量
// Gap 每一次请求的睡眠时间，限制吞吐量
var Glc = new(chan interface{})
var Gap = new(time.Duration)

// CreateGoQuery 所有的http请求都通过这里发送
func CreateGoQuery(urlStr string) (*goquery.Document, error) {
	// 并发限制
	*Glc <- 1
	defer func() {
		if *Gap > 0 {
			time.Sleep(*Gap)
		}
		_ = <-*Glc
	}()

	var resp string
	err := goutclient.GetClient().GET(urlStr).SetHeader(gout.H{
		"User-Agent": viper.GetString("common.useragent2"),
	}).BindBody(&resp).Do()
	if err != nil {
		return nil, err
	}
	if dom, err := goquery.NewDocumentFromReader(strings.NewReader(resp)); err != nil {
		log.Println("goquery.new", err)
		return nil, err
	} else {
		return dom, nil
	}
}

// CreateCrawler 暂时只生产两个类
func CreateCrawler(novelUrlStr string) (CrawlerInterface, error) {

	novelUrl, err := u.Parse(novelUrlStr)
	if err != nil {
		return nil, err
	}
	if _, ok := consts.BiQuGeInfoByHost[novelUrl.Hostname()]; ok {
		return &BiQuGeCrawler{
			novelUrl: novelUrl,
			filter:   &chapterFilterCommon{},
		}, nil
	}

	if _, ok := consts.NewSiteInfoConfigMap[novelUrl.Hostname()]; ok {
		return &NewBiQuGeCrawler{
			novelUrl:   novelUrl,
			nextGetter: &nextGetterCommon{},
		}, nil
	}
	return nil, errors.New("暂时不支持该网站")
}

// GbkToUtf8 GBK 转 UTF-8，如果本来就是UTF8那么本函数不进行任何操作
func GbkToUtf8(s []byte) ([]byte, error) {
	// 如果是uft8则直接返回
	if utf8.Valid(s) {
		return s, nil
	}
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := io.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func RemoveHtmlElem(str, selector string) (string, error) {

	dom, err := goquery.NewDocumentFromReader(strings.NewReader(str))
	if err != nil {
		return "", err
	}

	// 删除符合seletor的元素
	dom.Find(selector).Remove()

	res, err := dom.Html()
	if err != nil {
		return "", err
	}

	res = res[25 : len(res)-14]
	return res, nil
}
