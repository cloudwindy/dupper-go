package crawler

import (
	"io"
	"os"

	//    "fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

// Data structures
type Tag struct {
	Name  string
	Count int
	URL   string
}

type Comic struct {
	ID           int
	EnglishName  string
	OriginalName string
	Parodies     []string
	Characters   []string
	Tags         []string
	Artists      []string
	Groups       []string
	Translated   bool
	Language     []string
	Categories   string
	DownloadInfo
}

type DownloadInfo struct {
	DownloadID  int
	DownloadLan string
	DownloadNum int
	DownloadExt string
}

// Error(not nil) -> Panic
func check(err error) {
	if err != nil {
		panic(err)
	}
}

const PROXY_ADDR string = "http://127.0.0.1:1080"

//const PROXY_ADDR string = ""
const TIMEOUT = 30 * time.Second
const TOTAL_TIMEOUT = 2 * time.Minute
const WAIT = 2 * time.Second

var httpClient http.Client

func init() {
	if PROXY_ADDR != "" {
		proxy, err := url.Parse(PROXY_ADDR)
		check(err)
		netTransport := &http.Transport{
			Proxy:                 http.ProxyURL(proxy),
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second * time.Duration(5),
		}
		httpClient = http.Client{
			Timeout:   TIMEOUT,
			Transport: netTransport,
		}
	} else {
		httpClient = http.Client{}
	}
}

// 带错误重试的Get
func get(url string) (res *http.Response) {
	var err error
	deadline := time.Now().Add(TOTAL_TIMEOUT)
	for tries := 1; time.Now().Before(deadline); tries += 1 {
		res, err = httpClient.Get(url)
		if err == nil {
			return res
		}
		log.Errorf("%s 正在进行第 %d 次重试...", err, tries)
		time.Sleep(WAIT)
	}
	log.Fatalf("%s", err)
	return
}

// 获取HTML文档
func fetchDocument(urlContent string) (doc *goquery.Document) {
	res := get("https://a-upp.com/" + urlContent + "/")
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		log.Fatalf("获取文档时: %s", err)
	}
	return
}

// 保存文件
func saveFile(url string, path string) {
	res := get(url)
	defer res.Body.Close()
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Errorf("打开文件 %s 时: %s", path, err)
		return
	}
	defer file.Close()
	_, err = io.Copy(file, res.Body)
	if err != nil {
		log.Errorf("写入文件 %s 时: %s", path, err)
		return
	}
	return
}

// 转换到数字
func toInt(val string) (res int) {
	res, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("转换字符串 %s 到数字时: %s", val, err)
	}
	return
}

// 转换到字符串
func toStr(val int) (res string) {
	res = strconv.Itoa(val)
	return
}

// 获取所有标签
func FetchTags() (tags []Tag) {
	doc := fetchDocument("tags")
	tags = make([]Tag, doc.Find("a.tag").Length())
	doc.Find("a.tag").Each(func(idx int, tag *goquery.Selection) {
		// 获取标签数量
		count_s := tag.Find("span.count").Text()
		count := strings.TrimRight(strings.TrimLeft(count_s, "("), ")")
		// 获取后删除
		tag.Find("span.count").Remove()
		// 保存标签
		tags[idx].Name = tag.Text()
		tags[idx].Count = toInt(count)
		tags[idx].URL, _ = tag.Attr("href")
	})
	return
}

// 获取整页本子信息
func FetchPage(page int) (comic []Comic) {
	// 生成页面URL
	var url string
	if page > 0 {
		url += "page/" + toStr(page)
	}
	// 获取页面
	doc := fetchDocument(url)
	// 提取所有页面上本子的URL
	gallery := doc.Find(".cover")
	// 新建内存空间
	// IDs: 存储所有的本子id
	// comic: 存储所有的本子信息
	// chs: 从协程获取本子信息的管道
	var ids []int
	ch := make(chan Comic)
	// 预防内存泄漏
	defer close(ch)
	// 依次处理URL
	gallery.Each(func(idx int, sec *goquery.Selection) {
		// 获取URL
		idssss, _ := sec.Attr("href")
		// 把URL切断成四片
		idsss := strings.Split(idssss, "/")
		// 第三片就是ID
		idss := idsss[2]
		// 现在ID仍然是string 把他转换成int
		id := toInt(idss)
		// 加入数组
		ids = append(ids, id)
	})
	for _, id := range ids {
		go func(id int) { ch <- NewComic(id) }(id)
	}
	for _, _ = range ids {
		comic = append(comic, <-ch)
	}
	return
}

// 根据ID获取本子信息
func NewComic(comicId int) (comic Comic) {
	comic = Comic{}
	sDoc := fetchDocument("s/" + toStr(comicId))
	comic.ID = comicId
	comic.EnglishName = sDoc.Find("h1").Text()
	comic.OriginalName = sDoc.Find("h2").Text()
	sDoc.Find("a.tag").Each(func(_ int, sec *goquery.Selection) {
		tagURL, _ := sec.Attr("href")
		splited := strings.Split(tagURL, "/")
		tagType := splited[1]
		tagName := splited[2]
		switch tagType {
		case "parodies":
			comic.Parodies = append(comic.Parodies, tagName)
		case "tags":
			comic.Tags = append(comic.Tags, tagName)
		case "artists":
			comic.Artists = append(comic.Artists, tagName)
		case "groupss":
			comic.Groups = append(comic.Groups, tagName)
		case "languages":
			if tagName == "translated" {
				comic.Translated = true
			} else {
				comic.Language = append(comic.Language, tagName)
			}
		case "categories":
			comic.Categories = tagName
		}
	})
	gDoc := fetchDocument("g/" + toStr(comicId)).Find(".img-url:last-of-type")
	downloadURL := strings.Split(gDoc.Text(), "/")
	comic.DownloadLan = downloadURL[4]
	comic.DownloadID = toInt(downloadURL[5])
	downloadFileName := strings.Split(downloadURL[6], ".")
	comic.DownloadNum = toInt(downloadFileName[0])
	comic.DownloadExt = downloadFileName[1]
	return
}

// 下载本子
func (comic DownloadInfo) Download(dir string) {
	ch := make(chan int)
	for n := 1; n <= comic.DownloadNum; n++ {
		go func(n int) {
			fileName := toStr(n) + "." + comic.DownloadExt
			url := "https://a.comicstatic.icu/img/" + comic.DownloadLan + "/" + toStr(comic.DownloadID) + "/" + fileName
			saveFile(url, dir+"/"+fileName)
			ch <- 0
		}(n)
	}
	for n := 1; n <= comic.DownloadNum; n++ {
		<-ch
	}
}
