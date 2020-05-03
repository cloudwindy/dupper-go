package main

import (
	"fmt"
	"testing"
)

// TEST PASSED
func testFetchTags(t *testing.T) {
	tags := FetchTags()
	if len(tags) <= 0 {
		t.Errorf("No tag found")
		return
	}
	fmt.Println("Tags count:", len(tags))
	return
}

// TEST PASSED
func testNewComic(t *testing.T) {
	comic := NewComic(274506)
	comic.Info()
	return
}

// TEST PASSED
func TestDownload(t *testing.T) {
	comic := NewComic(295776)
	comic.Download("./save")
}

// TEST PASSED
func testFetchPage(t *testing.T) {
	comics := FetchPage(1)
	for _, comic := range comics {
		comic.Info()
	}
}

func (comic Comic) Info() {
	fmt.Println("本子 ID: ", comic.ID)
	fmt.Println("英文名称: ", comic.EnglishName)
	fmt.Println("原文名称: ", comic.OriginalName)
	fmt.Println("系列: ", comic.Parodies)
	fmt.Println("标签: ", comic.Tags)
	fmt.Println("作者: ", comic.Artists)
	fmt.Println("团体: ", comic.Groups)
	fmt.Println("已翻译: ", comic.Translated)
	fmt.Println("语言: ", comic.Language)
	fmt.Println("分类: ", comic.Categories)
	fmt.Println("下载 ID: ", comic.DownloadID)
	fmt.Println("下载页数:", comic.DownloadNum)
	fmt.Println("下载语言代码:", comic.DownloadLan)
	fmt.Println("下载文件后缀:", comic.DownloadExt)
	return
}
