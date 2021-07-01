package main

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"httpBruter/pkg/randHeader"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	t := make(chan bool, 10)
	start := time.Now()
	for i := 0; i < len(t); i++ {
		//cacel := QueueScreenShot("https://bilibili.com")
		//cacel()
		go func() {
			rodScreenshot()
			t <- true
		}()
		<-t
	}

	end := time.Now()
	fmt.Printf("总耗时:%d s", (end.UnixNano()-start.UnixNano())/1e9)

}

func QueueScreenShot(Url string) context.CancelFunc {

	options := []chromedp.ExecAllocatorOption{
		chromedp.WindowSize(1440, 900),
		chromedp.Flag("headless", true),
		chromedp.Flag("ignore-certificate-errors", true),
		// 禁用扩展
		chromedp.Flag("disable-extensions", true),
		// 禁止加载所有插件
		chromedp.Flag("disable-plugins", true),
		// 禁用浏览器应用
		chromedp.Flag("disable-software-rasterizer", true),
		// 禁用GPU，不显示GUI
		chromedp.DisableGPU,
		// 取消沙盒模式
		chromedp.NoSandbox,
		chromedp.UserAgent(randHeader.RandHeader()),
	}
	options = append(chromedp.DefaultExecAllocatorOptions[:], options...)
	parent, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	context1, cancel1 := chromedp.NewContext(parent)
	defer cancel1()
	context2, cancel2 := context.WithTimeout(context1, 20*time.Second)

	chromedp.ListenTarget(context2, func(ev interface{}) { //去除alert
		if _, ok := ev.(*page.EventJavascriptDialogOpening); ok {
			t := page.HandleJavaScriptDialog(false)
			go func() {
				if err := chromedp.Run(context2, t); err != nil {
					fmt.Println("拦截弹窗失败")
				}
				chromedp.Click("#alert", chromedp.ByID)
			}()
		}
	})

	captureScreenshot := make([]byte, 0)
	_ = chromedp.Run(context2,
		chromedp.Navigate(Url),
		chromedp.CaptureScreenshot(&captureScreenshot),
	)
	saveimg(captureScreenshot, Url)
	return cancel2
}

/*保存图片*/
func saveimg(buf []byte, url string) {
	//组合path
	var path string
	p, _ := os.Getwd()
	filename := filepath.Join(p, "testScreenshot")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err := os.Mkdir(filename, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
	path1 := "" //协议类型
	path2 := "" //domain
	path3 := "" //端口号，如果没有端口号，则默认80，或443
	//defaultURL
	if strings.Contains(url, "http://") {
		path1 = "http"
	} else {
		path1 = "https"
	}

	path2 = strings.Replace(url, "http://", "", -1)
	path2 = strings.Replace(path2, "https://", "", -1)

	if strings.Contains(path2, ":") {
		path3 = strings.Split(path2, ":")[1]
	} else {
		if path1 == "http" {
			path3 = "443"
		} else {
			path3 = "443"
		}
	}
	path = path1 + "-" + path2 + "-" + path3 + ".jpg"
	path = filepath.Join(filename, path)

	err := ioutil.WriteFile(path, buf, 0o644)
	if err != nil {
		log.Fatal(err)
	}
}

// This example demonstrates how to take a screenshot of a specific element and
// of the entire browser viewport, as well as using `kit`
// to store it into a file.
func rodScreenshot() {
	browser := rod.New().MustConnect()
	defer browser.Close()

	//capture entire browser viewport, returning jpg with quality=90
	buf, err := browser.MustPage("https://www.bilibili.com").Screenshot(true, &proto.PageCaptureScreenshot{
		Format:  proto.PageCaptureScreenshotFormatJpeg,
		Quality: 90,
	})
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("bilibili.png", buf, 0644)
	if err != nil {
		panic(err)
	}
}
