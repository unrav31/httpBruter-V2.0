package main

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"httpBruter/pkg/randHeader"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func main() {
	//chrome("https://www.bilibili.com")
	chrome("https://weibo.com/login.php")
}

func chrome(url string) {
	opt := []chromedp.ExecAllocatorOption{ //设置参数
		chromedp.WindowSize(1440, 900),
		chromedp.Flag("headless", true),
		// 禁用扩展
		chromedp.Flag("disable-extensions", true),
		// 禁止加载所有插件
		chromedp.Flag("disable-plugins", true),
		// 禁用浏览器应用
		chromedp.Flag("disable-software-rasterizer", true),
		// 隐身模式启动
		chromedp.Flag("incognito", true),
		// 取消沙盒模式
		chromedp.NoSandbox,
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.DisableGPU,
		chromedp.UserAgent(randHeader.RandHeader()),
	}

	opt = append(chromedp.DefaultExecAllocatorOptions[:], opt...)
	parent, cancel1 := chromedp.NewExecAllocator(context.Background(), opt...)
	defer cancel1()
	context1, cancel2 := chromedp.NewContext(parent)
	defer cancel2()
	context2, cancel3 := context.WithTimeout(context1, 30*time.Second)
	defer cancel3()

	chromedp.ListenTarget(context2, func(ev interface{}) { //拦截弹窗
		if _, ok := ev.(*page.EventJavascriptDialogOpening); ok {
			t := page.HandleJavaScriptDialog(true)
			go func() {
				if err := chromedp.Run(context2, t); err != nil {
					fmt.Println("拦截弹窗失败")
				}
				chromedp.Click("#alert", chromedp.ByID)
			}()
		}
	})

	var title, html, location string

	fullScreenshot := make([]byte, 0)
	_, err := chromedp.RunResponse(context2,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`html`, chromedp.ByQuery),
		chromedp.Title(&title),
		chromedp.OuterHTML(`html`, &html, chromedp.ByQuery),
		chromedp.Location(&location),
		chromedp.ActionFunc(func(ctx context.Context) error {
			for i := 0; i < 1000000; i += 900 {
				_, exp, err := runtime.Evaluate(fmt.Sprintf(`window.scrollTo(0,%d);`, i)).Do(ctx)
				time.Sleep(500)
				if err != nil {
					log.Fatal(err)
				}
				if exp != nil {
					log.Fatal(exp)
				}
			}
			return nil
		}),
		chromedp.FullScreenshot(&fullScreenshot, 90),
	)

	if err != nil {
		log.Fatal(err)
	}

	err2 := ioutil.WriteFile("weibo.jpg", fullScreenshot, os.ModePerm)
	if err2 != nil {
		return
	}
}
