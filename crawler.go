package main

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

// DownloadInfo 下载信息
type DownloadInfo struct {
	ID  int
	Lan string
	Num int
	Ext string
}

// Error(not nil) -> Panic
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// 反屏蔽

const SITE string = "d-upp.net"
//const PROXY_ADDR string = "http://127.0.0.1:1080"
const PROXY_ADDR string = ""

const TIMEOUT = 30 * time.Second
const MAXRETRY = 2 * time.Minute
const WAIT = 2 * time.Second

var httpClient http.Client

func init() {
	if PROXY_ADDR != "" {
		proxy, err := url.Parse(PROXY_ADDR)
		check(err)
		netTransport := &http.Transport{
			Proxy: http.ProxyURL(proxy),
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

// 错误重试
func autoRetry(try func(string) (*http.Response, error), params string) (res *http.Response) {
	var err error
	deadline := time.Now().Add(MAXRETRY)
	for tries := 1; time.Now().Before(deadline); tries++ {
		res, err = try(params)
		if err == nil {
			return res
		}
		log.Errorf("%s 正在进行第 %d 次重试...", err, tries)
		time.Sleep(WAIT)
	}
	log.Fatalf("%s", err)
	panic(err)
	return
}

// 获取HTML文档
func fetchDocument(method func(string) (*http.Response, error), urlContent string) (doc *goquery.Document) {
	res := autoRetry(method, "https://" + SITE + "/" + urlContent + "/")
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		log.Fatalf("获取文档时: %s", err)
	}
	return
}

// 保存文件
func saveFile(url string, path string) {
	res := autoRetry(httpClient.Get, url)
	defer res.Body.Close()
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
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
func toInt(str string) (num int) {
	num, err := strconv.Atoi(str)
	if err != nil {
		log.Fatalf("转换字符串 %s 到数字时: %s", str, err)
	}
	return
}

// 转换到字符串
func toStr(num int) (str string) {
	str = strconv.Itoa(num)
	return
}

// GetDownloadInfo 根据ID获取本子信息
func GetDownloadInfo(comicID int) (info DownloadInfo) {
	info = DownloadInfo{}
	gDoc := fetchDocument(httpClient.Get, "g/" + toStr(comicID)).Find(".img-url:last-of-type")
	URL := strings.Split(gDoc.Text(), "/")
	info.Lan = URL[4]
	info.ID = toInt(URL[5])
	FileName := strings.Split(URL[6], ".")
	info.Num = toInt(FileName[0])
	info.Ext = FileName[1]
	return
}

// Download 下载本子
func (info DownloadInfo) Download(dir string) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		log.Info("保存路径不存在 已新建文件夹")
		os.MkdirAll(dir, 0755)
	}
	ch := make(chan int)
	for n := 1; n <= info.DownloadNum; n++ {
		go func(n int, info DownloadInfo, ch chan) {
			fileName := toStr(n) + "." + info.Ext
			url := "https://a.comicstatic.icu/img/" + info.Lan + "/" + toStr(info.ID) + "/" + fileName
			saveFile(url, dir + "/" + fileName)
			ch <- 0
		}(n, info, ch)
	}
	for n := 1; n <= comic.DownloadNum; n++ {
		<-ch
	}
}
