package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	core "twist/core/json"
)

var (
	ua = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.116 Safari/537.36`
)

// StoreRet
type Data struct {
	Html  string   `json:"html"`
	Title string   `json:"title"`
	Imgs  []string `json:"imgs"`
}

type Server struct {
	Addr string
}

// Response
type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data *Data  `json:"data"`
}

type Req struct {
	Link string `json:"link"`
}

func Start(addr string) {
	var s = &Server{Addr: addr}
	go func() {
		var (
			err      error
			serveMux = http.NewServeMux()
		)
		serveMux.HandleFunc("/post", s.Post)
		if err = http.ListenAndServe(fmt.Sprintf(":%s", s.Addr), serveMux); err != nil {
			log.Fatalf("http.ListenAndServe(\"%s\") error(%v)", addr, err)
			return
		}
	}()
}

func (s *Server) Post(w http.ResponseWriter, r *http.Request) {
	rsp := &Response{
		Code: -1,
		Msg:  "",
	}
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rsp.Msg = err.Error()
		JSONResp(r, w, rsp)
		return
	}
	req := &Req{}
	err = core.JSON.Unmarshal(raw, &req)
	if err != nil {
		rsp.Msg = err.Error()
		JSONResp(r, w, rsp)
		return
	}
	link := strings.TrimSpace(req.Link)
	if link == "" {
		rsp.Msg = "链接错误"
		JSONResp(r, w, rsp)
		return
	}
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true),
		chromedp.UserAgent(ua),
	}

	c, _ := chromedp.NewExecAllocator(context.Background(), options...)

	chromeCtx, cancel := chromedp.NewContext(c, chromedp.WithLogf(log.Printf))
	_ = chromedp.Run(chromeCtx, make([]chromedp.Action, 0, 1)...)

	timeOutCtx, cancel := context.WithTimeout(chromeCtx, 60*time.Second)
	defer cancel()

	var htmlcon string
	err = chromedp.Run(timeOutCtx,
		chromedp.Navigate(link),
		//等待某个特定的元素出现
		chromedp.OuterHTML(`document.querySelector("html")`, &htmlcon, chromedp.ByJSPath),
		//需要爬取的网页的url
		//chromedp.WaitVisible(`div[id="J_DivItemDesc"]`),
		//chromedp.WaitVisible(`div[class="water-container"]`),
		//生成最终的html文件并保存在htmlContent文件中
	)
	if err != nil {
		log.Fatal(err)
		rsp.Msg = err.Error()
		JSONResp(r, w, rsp)
		return
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer([]byte(htmlcon)))
	if err != nil {
		log.Fatal(err)
		rsp.Msg = err.Error()
		JSONResp(r, w, rsp)
		return
	}
	imgs := make([]string, 0)
	doc.Find("img").Each(func(i int, sel *goquery.Selection) {
		img, _ := sel.Attr("src")
		if img != "" {
			img = strings.TrimLeft(img, "//")
			if img[0:4] != "http" {
				img = fmt.Sprintf("http://%s", img)
			}
			imgs = append(imgs, img)
		}
		dsrc, _ := sel.Attr("data-src")
		if dsrc != "" {
			dsrc = strings.TrimLeft(dsrc, "//")
			if dsrc[0:4] != "http" {
				dsrc = fmt.Sprintf("http://%s", dsrc)
			}
			imgs = append(imgs, dsrc)
		}
	})
	//first := doc.Find("h3.tb-main-title").First()
	first := doc.Find("title").First()
	rsp.Code = 0
	rsp.Data = &Data{
		Html:  htmlcon,
		Imgs:  imgs,
		Title: strings.TrimSpace(first.Text()),
	}

	JSONResp(r, w, rsp)
	return
}

func JSONResp(r *http.Request, wr http.ResponseWriter, res *Response) {
	var (
		err      error
		byteJson []byte
	)
	if byteJson, err = core.JSON.Marshal(res); err != nil {
		log.Fatalf("json.Marshal(\"%v\") failed (%v)", res, err)
		return
	}
	wr.Header().Set("Content-Type", "application/json;charset=utf-8")
	if _, err = wr.Write(byteJson); err != nil {
		log.Fatalf("HttpWriter Write error(%v)", err)
		return
	}
	start := time.Now()
	log.Printf("%s path:%s(params:%s,time:%f,ret:%v)", r.Method,
		r.URL.Path, r.Form.Encode(), time.Now().Sub(start).Seconds(), res)
}

func Signal() {
	var (
		c chan os.Signal
		s os.Signal
	)
	c = make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP)
	for {
		s = <-c
		log.Printf("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			return
		case syscall.SIGHUP:
		default:
			return

		}
	}
}

func main() {
	Start("8410")
	log.Println("init http api....")
	Signal()
}
