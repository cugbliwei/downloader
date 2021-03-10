package main

import (
	"io/ioutil"
	"net/http"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type Downloader struct {
	Client      *resty.Client
	Request     *resty.Request
	crawlerName string
	sessionId   string
	filedir     string
	filename    string
	isWrite     bool
	isMock      bool
}

func NewDownloader() *Downloader {
	return &Downloader{
		Client: resty.New(),
	}
}

func (hc *Downloader) SetCrawlerName(crawlerName string) {
	hc.crawlerName = crawlerName
}

func (hc *Downloader) SetSessionId(sessionId string) {
	hc.sessionId = sessionId
}

func (hc *Downloader) SetFiledir(filedir string) {
	hc.filedir = filedir
}

func (hc *Downloader) SetFilename(filename string) {
	hc.filename = filename
}

func (hc *Downloader) SetIsWrite(isWrite bool) {
	hc.isWrite = isWrite
}

func (hc *Downloader) SetIsMock(isMock bool) {
	hc.isMock = isMock
}

func (hc *Downloader) SetAfterResponseWriteFile() {
	hc.Client = hc.Client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		// 初始化request，以防影响下一次的请求
		hc.Request = nil
		// 当为mock请求时不用保存源码
		if hc.isMock {
			return nil
		}

		if !hc.isWrite {
			log.Errorf("%s 未设置写入文件", hc.sessionId)
			return nil
		}

		if hc.crawlerName == "" || hc.sessionId == "" || hc.filedir == "" || hc.filename == "" {
			log.Errorf("%s 参数未设置，crawlerName: %s, sessionId: %s, filedir: %s, filename: %s", hc.crawlerName, hc.sessionId, hc.filedir, hc.filename)
		}
		log.Infof("%s 源码写入文件: %s，内容长度：%d", hc.sessionId, hc.filename, len(resp.Body()))
		ioutil.WriteFile(hc.filedir+hc.filename, resp.Body(), 0644)
		return nil
	})
}

func (hc *Downloader) request(method, link string) (*resty.Response, []byte, error) {
	if hc.isMock {
		log.Infof("%s 从文件: %s 中获取源码", hc.sessionId, hc.filename)
		body, err := ioutil.ReadFile(hc.filedir + hc.filename)
		if err != nil {
			return nil, body, err
		}

		resp := &resty.Response{}
		resp.RawResponse = &http.Response{
			StatusCode: 200,
		}
		return resp, body, nil
	}

	if hc.Request != nil {
		resp, err := hc.Request.Get(link)
		return resp, resp.Body(), err
	}
	resp, err := hc.Client.R().Execute(method, link)
	return resp, resp.Body(), err
}

func (hc *Downloader) Get(link string) (*resty.Response, []byte, error) {
	return hc.request("GET", link)
}

func (hc *Downloader) Post(link string) (*resty.Response, []byte, error) {
	return hc.request("POST", link)
}

func main() {
	downloader := NewDownloader()
	downloader.SetCrawlerName("baidu")
	downloader.SetSessionId("123456")

	// 请求163页面，并保存源码
	downloader.SetIsWrite(true)
	downloader.SetFiledir("./")
	downloader.SetFilename("163.html")
	downloader.SetAfterResponseWriteFile()
	_, body, err := downloader.Get("https://www.163.com/")
	if err != nil {
		log.Error(err)
	}
	log.Infof("163 body length: %d", len(body))

	// 请求163页面，从文件中获取源码
	downloader.SetIsMock(true)
	downloader.SetFilename("163.html")
	downloader.Request = downloader.Client.SetRetryCount(2).R().SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.192 Safari/537.36")
	_, body, err = downloader.Get("https://www.163.com/")
	if err != nil {
		log.Error(err)
	}
	log.Infof("163 body length: %d", len(body))
}
