package retryable

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"httpBruter/pkg/decode"
	"httpBruter/pkg/errors"
	"httpBruter/pkg/options"
	"httpBruter/pkg/randHeader"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptrace"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Params struct {
	RAW             string // 存放从文件中读取的原始URL
	DefaultURL      string // 存放默认请求的URL
	PortURL         string // 如果是port访问，在这里存放，并以port进行访问，如果port带443，则将80端口放在Fallback中
	HTTPSURL        string // 存放HTTPSURL，用于标记是否为https
	HTTPURL         string // 存放HTTPURL，用于标记是否为http
	FallbackURL     string // 存放fallback的URL，必定是80端口
	RetryTimes      int
	Timeout         time.Duration
	Cookie          *http.Cookie
	Redirects       bool
	UserAgent       string
	Refer           string
	HeadlessTimeout time.Duration
}

type Response struct {
	TrueURL      string //实际请求的URL
	NetConn      httptrace.GotConnInfo
	ResponseTime int64
	Request      *http.Request
	ResponseList []*http.Response
	ResponseText []string //已经解码过的文本

	HeadlessResponse     *network.Response
	HeadlessTitle        string
	HeadlessLocation     string
	HeadlessResponseTime string
	HeadlessText         string

	Error         error
	HttpCanAccess bool //http是否能访问到

	ScreenshotBase string //图片的base64编码
}

// Request 请求逻辑函数
func (p *Params) Request(URL string, responseList []*Response, args *options.Args) []*Response {

	response := &Response{}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	jar, _ := cookiejar.New(nil)

	//构造请求参数和超时信息
	client := &http.Client{Timeout: p.Timeout * time.Second, Jar: jar}

	//如果有不跟随重定向选项
	if args.Redirects {
		client = &http.Client{Timeout: p.Timeout * time.Second, Jar: jar, CheckRedirect: noRedirect}
	}

	//请求结束后关闭连接
	defer client.CloseIdleConnections()

	request, reqErr := http.NewRequest("GET", URL, nil)
	response.TrueURL = URL
	//构造请求失败，经过Retry处理
	if reqErr != nil {
		retryReseponse := p.Retry(URL, reqErr, args)
		if retryReseponse == nil {
			return nil
		} else {
			response = retryReseponse.(*Response)
		}

	}

	//回调函数，获取服务器ip地址
	trace := &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			response.NetConn = connInfo
		},
	}

	request = request.WithContext(httptrace.WithClientTrace(request.Context(), trace))
	if args.Clistats != nil {
		args.Clistats.IncrementCounter("Request", 1)
	}
	request.Header.Add("User-Agent", p.UserAgent)
	request.Header.Add("referer", p.Refer)
	//设置cookie，用于获取shiro的指纹信息
	if p.Cookie != nil {
		request.AddCookie(p.Cookie)
	}
	response.Request = request

	start := time.Now()
	res, resErr := client.Do(request)

	//获取响应失败了,交给Retry去处理
	if resErr != nil {
		retryReseponse := p.Retry(URL, resErr, args)
		if retryReseponse == nil {
			return nil
		} else {
			response = retryReseponse.(*Response)
		}
		//如果重试后还是空，直接返回了

	}
	//页面响应时间
	response.ResponseTime = (time.Now().UnixNano() - start.UnixNano()) / 1e6

	//包含重定位的响应列表
	response.ResponseList = redirections(res)

	responseList = append(responseList, response)
	//判断http是否能访问到
	for _, httpResponse := range response.ResponseList {
		if strings.Contains(httpResponse.Request.URL.String(), "http://") {
			response.HttpCanAccess = true
			break
		} else {
			continue
		}
	}
	return responseList
}

//重定向处理，不跟随重定向
func noRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

//重定向历史记录
func redirections(resp *http.Response) []*http.Response {
	history := make([]*http.Response, 0)
	for resp != nil {
		req := resp.Request
		history = append(history, resp)
		resp = req.Response
	}

	return history
}

//判断是ip还是domain
func isIP(url string) bool {
	URLPattern := `[a-zA-Z]+`
	reg := regexp.MustCompile(URLPattern)
	if !reg.MatchString(url) {
		return true

	} else {
		return false
	}
}

// MakeURL 构造即将请求的URL，填入对应的URL信息
func (p *Params) MakeURL(RAW string, args *options.Args) {
	p.RAW = RAW

	//包含http头部或https头部的情况
	if strings.Contains(RAW, "https://") {
		p.DefaultURL = RAW
		p.HTTPSURL = RAW
		p.HTTPURL = strings.Replace(RAW, "https://", "http://", -1)
		//并包含ip的情况
		if isIP(RAW) {
			reg := regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)`)
			p.PortURL = reg.FindString(RAW)
		}
		//并包含443端口，而且开启了fallback参数
		if strings.Contains(RAW, ":443") && args.NoFallback {
			p.FallbackURL = strings.Replace(p.DefaultURL, "https://", "http://", -1)
			p.FallbackURL = strings.Replace(p.FallbackURL, ":443", ":80", -1)
		}

	} else if strings.Contains(RAW, "http://") {
		p.DefaultURL = RAW
		p.HTTPURL = RAW
		//并包含ip的情况
		if isIP(RAW) {
			reg := regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)`)
			p.PortURL = reg.FindString(RAW)
		}

		//不包含http头部，但包含端口号，查看是443或者80端口，拼接对应的头部
	} else if strings.Contains(RAW, ":") {
		//是否包含ip
		if isIP(RAW) {
			reg := regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)`)
			p.PortURL = reg.FindString(RAW)
		}
		//包含443端口
		if RAW[len(RAW)-3:] == "443" {
			p.DefaultURL = "https://" + RAW
			p.HTTPSURL = "https://" + RAW
			p.FallbackURL = func(f bool) string {
				if f {
					return "http://" + strings.Replace(RAW, ":443", ":80", -1)
				} else {
					return ""
				}
			}(args.NoFallback)

			//本来就带80端口的
		} else if RAW[len(RAW)-3:] == ":80" {
			p.DefaultURL = "http://" + RAW
			p.HTTPURL = "http://" + RAW

			//其他端口号
		} else {
			p.DefaultURL = "https://" + RAW
			p.HTTPSURL = "https://" + RAW
		}

		//不包含头部，也不包含端口号，默认https，返回给HTTPSURL，并设置fallbackurl为80端口
	} else {
		if isIP(RAW) {
			reg := regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+)`)
			p.PortURL = reg.FindString(RAW)
		}
		p.DefaultURL = "https://" + RAW + ":443"
		p.HTTPSURL = "https://" + RAW + ":443"
		p.HTTPURL = "http://" + RAW + ":80"
		p.FallbackURL = func(f bool) string {
			if f {
				return "http://" + RAW + ":80"
			} else {
				return ""
			}
		}(args.NoFallback)
	}

}

// Retry 重试逻辑函数
func (p *Params) Retry(URL string, err error, args *options.Args) interface{} {

	if strings.Contains(err.Error(), "HTTPS client") {
		url := strings.Replace(p.DefaultURL, "https://", "http://", -1)
		url = strings.Replace(url, ":443", ":80", -1)
		retryResponse := p.retryable(url)
		if retryResponse != nil {
			return retryResponse.(*Response)
		}
	} else if strings.Contains(err.Error(), "Timeout") && p.RetryTimes != 0 {
		for i := 0; i < p.RetryTimes; i++ {
			retryResponse := p.retryable(URL)
			if retryResponse != nil {
				return retryResponse.(*Response)
			} else {
				continue
			}
		}
	} else {
		if args.Clistats != nil {
			//统计一条错误
			args.Clistats.IncrementCounter("Error", 1)
		}
		errors.HandleRequestErrors(p.DefaultURL, err, args.Debug)
	}
	return nil
}

func (p *Params) retryable(URL string) interface{} {
	response := &Response{}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	jar, _ := cookiejar.New(nil)

	//构造请求参数和超时信息
	client := &http.Client{Timeout: p.Timeout * time.Second, Jar: jar}

	//如果有不跟随重定向选项
	if p.Redirects {
		client = &http.Client{Timeout: p.Timeout * time.Second, Jar: jar, CheckRedirect: noRedirect}
	}

	//请求结束后关闭连接
	defer client.CloseIdleConnections()

	request, reqErr := http.NewRequest("GET", URL, nil)
	response.TrueURL = URL
	//构造请求失败，返回空
	if reqErr != nil {
		return nil
	}

	//回调函数，获取服务器ip地址
	trace := &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			response.NetConn = connInfo
		},
	}

	request = request.WithContext(httptrace.WithClientTrace(request.Context(), trace))
	request.Header.Add("User-Agent", p.UserAgent)
	request.Header.Add("referer", p.Refer)
	//设置cookie，用于获取shiro的指纹信息
	if p.Cookie != nil {
		request.AddCookie(p.Cookie)
	}
	response.Request = request

	start := time.Now()
	res, resErr := client.Do(request)

	//获取响应失败了,返回空
	if resErr != nil {
		return nil
	}
	//页面响应时间
	response.ResponseTime = (time.Now().UnixNano() - start.UnixNano()) / 1e6

	//包含重定位的响应列表
	response.ResponseList = redirections(res)

	//判断http是否能访问到
	if len(response.ResponseList) != 0 && p.HTTPURL != "" {
		response.HttpCanAccess = true
	}
	return response
}

// ToText 获取响应文本到Response结构体中
func (r *Response) ToText() {
	for _, response := range r.ResponseList {
		readBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			continue
		}
		decodeBytes := decode.MainParse(readBytes)
		r.ResponseText = append(r.ResponseText, string(decodeBytes))
	}
}

func (p *Params) Chromedp(args *options.Args, url string, response *Response) {
	opt := []chromedp.ExecAllocatorOption{ //设置参数
		chromedp.ExecPath(args.Binary),
		chromedp.WindowSize(1440, 900),
		chromedp.Flag("headless", true),
		//chromedp.Flag("blink-settings", "imageEnabled=false"),
		//chromedp.Flag("disable-images", true),
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

	var title, html, location, responseTime string

	fullScreenshot := make([]byte, 0)
	responseTimeInt := time.Now().UnixNano()
	resp, err := chromedp.RunResponse(context2,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`html`, chromedp.ByQuery),
		//chromedp.Sleep(1*time.Second), //等待js加载完成，再获取title
		//chromedp.WaitSelected(`head > title`,chromedp.ByQuery),
		chromedp.Title(&title),
		chromedp.OuterHTML(`html`, &html, chromedp.ByQuery),
		chromedp.Location(&location),
		//如果有截图参数，就返回截图，如果没有则返回一个空的task列表不做任何处理
		func(s bool) chromedp.EmulateAction {
			if s {
				if args.Clistats != nil {
					//统计一条截图
					args.Clistats.IncrementCounter("Screenshot", 1)
				}
				if args.ScreenshotFullPage {
					return chromedp.FullScreenshot(&fullScreenshot, 90)
				}
				return chromedp.CaptureScreenshot(&fullScreenshot)
			} else {
				return chromedp.Tasks{}
			}
		}(args.Screenshot),
	)
	if err != nil || resp == nil {
		if args.Clistats != nil {
			//统计一条错误
			args.Clistats.IncrementCounter("Error", 1)
		}
		errors.HandleRequestErrors(p.DefaultURL, err, args.Debug)
	}

	if args.Screenshot {
		saveImg(fullScreenshot, url, args.ScreenshotPath)
		if len(fullScreenshot) != 0 {
			//暂时先不存图片到数据库
			//response.ScreenshotBase = ScreenshotBase64(fullScreenshot)

			response.ScreenshotBase = ""
		} else {
		}
	}

	responseTimeInt = time.Now().UnixNano() - responseTimeInt

	responseTime = fmt.Sprintf("%d ms", responseTimeInt/1e6-1000)
	response.HeadlessTitle = func(c string, t string) string {
		if t == "" || strings.Count(t, "") < 5 {
			return t + " (" + h1Title(c) + ")"
		} else {
			return t
		}
	}(html, title)
	response.HeadlessResponse = resp
	response.HeadlessLocation = location
	response.HeadlessResponseTime = responseTime
	//关闭浏览器实例
	c := chromedp.FromContext(context2)
	defer browser.Close().Do(cdp.WithExecutor(context2, c.Browser))
}

// H1 匹配h1作为title
const h1 = `<h1>(.*?)</h1>|<H1>(.*?)</H1>`

func h1Title(content string) string {
	Reg := regexp.MustCompile(h1)
	h1title := Reg.FindString(content)
	h1title = strings.Replace(h1title, "<h1>", "", -1)
	h1title = strings.Replace(h1title, "<H1>", "", -1)
	h1title = strings.Replace(h1title, "</h1>", "", -1)
	h1title = strings.Replace(h1title, "</H1>", "", -1)
	return h1title
}

/*保存图片*/
func saveImg(buf []byte, url string, argPath string) {
	//组合path
	var path string
	var filename string

	if argPath != "" {
		filename = argPath
	} else {
		pp, _ := os.Getwd()
		filename = filepath.Join(pp, "screenshots")
	}

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
		split := strings.Split(path2, ":")
		path2 = split[0]
		path3 = split[1]
	} else {
		if path1 == "http" {
			path3 = "80"
		} else {
			path3 = "443"
		}
	}

	path = path1 + "-" + path2 + "-" + path3 + ".jpg"
	path = strings.Replace(path, "\\", "", -1)
	path = strings.Replace(path, "/", "", -1)
	path = filepath.Join(filename, path)

	if len(buf) != 0 {
		fileBuffer, err := os.Create(path)
		defer func(fileBuffer *os.File) {
			err := fileBuffer.Close()
			if err != nil {
				log.Fatal("[x] 截图后关闭图片失败")
			}
		}(fileBuffer)

		if err != nil {
			log.Fatal(err)
		}
		_, err2 := fileBuffer.Write(buf)
		if err2 != nil {
			return
		}

	} else {
		//截图链接超时
	}
}

func ScreenshotBase64(image []byte) string {
	encoded := base64.StdEncoding.EncodeToString(image)
	var buffer bytes.Buffer
	for i := 0; i < len(encoded); i++ {
		ch := encoded[i]
		buffer.WriteByte(ch)
		if (i+1)%76 == 0 {
			buffer.WriteByte('\n')
		}
	}
	buffer.WriteByte('\n')
	value := buffer.Bytes()
	return `<img src="data:image/png;base64,` + string(value) + `"/>`
}
