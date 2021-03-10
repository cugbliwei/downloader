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

func (d *Downloader) SetCrawlerName(crawlerName string) {
	d.crawlerName = crawlerName
}

func (d *Downloader) SetSessionId(sessionId string) {
	d.sessionId = sessionId
}

func (d *Downloader) SetFiledir(filedir string) {
	d.filedir = filedir
}

func (d *Downloader) SetFilename(filename string) {
	d.filename = filename
}

func (d *Downloader) SetIsWrite(isWrite bool) {
	d.isWrite = isWrite
}

func (d *Downloader) SetIsMock(isMock bool) {
	d.isMock = isMock
}

func (d *Downloader) SetAfterResponseWriteFile() {
	d.Client = d.Client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		// 初始化request，以防影响下一次的请求
		d.Request = nil
		// 当为mock请求时不用保存源码
		if d.isMock {
			return nil
		}

		if !d.isWrite {
			log.Errorf("%s 未设置写入文件", d.sessionId)
			return nil
		}

		if d.crawlerName == "" || d.sessionId == "" || d.filedir == "" || d.filename == "" {
			log.Errorf("%s 参数未设置，crawlerName: %s, sessionId: %s, filedir: %s, filename: %s", d.crawlerName, d.sessionId, d.filedir, d.filename)
		}
		log.Infof("%s 源码写入文件: %s，内容长度：%d", d.sessionId, d.filename, len(resp.Body()))
		ioutil.WriteFile(d.filedir+d.filename, resp.Body(), 0644)
		return nil
	})
}

func (d *Downloader) request(method, link string) (*resty.Response, []byte, error) {
	if d.isMock {
		log.Infof("%s 从文件: %s 中获取源码", d.sessionId, d.filename)
		body, err := ioutil.ReadFile(d.filedir + d.filename)
		if err != nil {
			return nil, body, err
		}

		resp := &resty.Response{}
		resp.RawResponse = &http.Response{
			StatusCode: 200,
		}
		return resp, body, nil
	}

	if d.Request != nil {
		resp, err := d.Request.Get(link)
		return resp, resp.Body(), err
	}
	resp, err := d.Client.R().Execute(method, link)
	return resp, resp.Body(), err
}

func (d *Downloader) Get(link string) (*resty.Response, []byte, error) {
	return d.request("GET", link)
}

func (d *Downloader) Post(link string) (*resty.Response, []byte, error) {
	return d.request("POST", link)
}

func main() {
	downloader := NewDownloader()
	downloader.SetCrawlerName("baidu")
	downloader.SetSessionId("829122b5-1514-4662-86fb-95f55c17f263")

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
