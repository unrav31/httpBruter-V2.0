package domain

import (
	"encoding/json"
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"httpBruter/pkg/options"
	"httpBruter/pkg/retryable"
	"httpBruter/pkg/searchCIDR"
	"httpBruter/pkg/structs"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// ReverseIP 反查IP
func ReverseIP(responseList []*retryable.Response, ua string, arg *options.Args) []structs.Domains {

	remoteAddr := ""
	for _, response := range responseList {
		remoteAddr = response.NetConn.Conn.RemoteAddr().String()
		break
	}

	reverseDomainList := make([]structs.Domains, 0)
	if searchCIDR.ReverseCIDR(remoteAddr, arg.CIDRMap) {
		return OpenBrowser(remoteAddr, ua, reverseDomainList, arg)
	} else {
		return reverseDomainList
	}

}

// OpenBrowser 开启一个浏览器页面并访问ipchaxun.com
func OpenBrowser(remoteAddr string, ua string, domains []structs.Domains, arg *options.Args) []structs.Domains {

	URL := "https://ipchaxun.com/"
	ip := strings.Split(remoteAddr, ":")[0]

	header := ua

	//这句代码是设置headless，设置true 或者false，这里要设置NewUserMode才能控制浏览器关闭
	u := launcher.NewUserMode().Headless(true).MustLaunch()

	req := proto.NetworkSetUserAgentOverride{UserAgent: header}
	browser := rod.New().ControlURL(u).MustConnect()
	defer func(browser *rod.Browser) {
		err := browser.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(browser)

	//修改请求头
	router := browser.HijackRequests()
	defer router.MustStop()
	router.MustAdd("*.js", func(ctx *rod.Hijack) {
		ctx.Request.Req().Header.Add("User-Agent", header)
		ctx.Request.Req().Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
		ctx.Request.Req().Header.Add("Accept-Encoding", "gzip, deflate, br")
		ctx.Request.Req().Header.Add("Accept-Language", "zh-CN,zh;q=0.9")
		ctx.Request.Req().Header.Add("Cache-Control", "max-age=0")
		ctx.Request.Req().Header.Add("Connection", "keep-alive")
		ctx.Request.Req().Header.Add("Host", "ipchaxun.com")
		//ctx.Request.Req().Header.Add("sec-ch-ua", "Not;A Brand\";v=\"99\", \"Google Chrome\";v=\"91\", \"Chromium\";v=\"91\"")
		ctx.Request.Req().Header.Add("sec-ch-ua-mobile", "?0")
		ctx.Request.Req().Header.Add("Sec-Fetch-Dest", "document")
		ctx.Request.Req().Header.Add("Sec-Fetch-Mode", "navigate")
		ctx.Request.Req().Header.Add("Sec-Fetch-Site", "none")
		ctx.Request.Req().Header.Add("Sec-Fetch-User", "?1")
		ctx.Request.Req().Header.Add("Upgrade-Insecure-Requests", "1")
	})
	go router.Run()


	opt := proto.TargetCreateTarget{URL: URL}

	page, err := browser.Page(opt)
	if err != nil {
		fmt.Println("[x] 请求ipchaxun.com失败，睡眠10秒后重试")
		time.Sleep(10 * time.Second)
		return OpenBrowser(remoteAddr, ua, domains, arg)
	}

	page.MustSetUserAgent(&req)
	page.MustWaitLoad()

	//以前是page.close，会阻塞，改成这个
	defer page.MustClose()

	page.MustNavigate(URL)//.Timeout(30 * time.Second)

	//设置鼠标滚动,滚动到元素可见（这一步没什么用，设置其他的焦点也可以，重点在下面的滚动步骤）
	el := page.MustElement("body > div > div.footer > div.mod-foot > div > div > p > span:nth-child(6)")
	el.MustFocus()
	mouse := page.Mouse

	//输入ip
	page.MustElement(`body > div > div.header > div > div.mod-head > div.bd > div > div.panels > div > div.input-box > input`).MustSelectAllText().MustInput(ip)
	//点击查询
	page.MustElement("body > div > div.header > div > div.mod-head > div.bd > div > div.panels > div > div.input-box > button.btn-search").MustClick()
	page.MustElement("#J_domain").MustWaitLoad()
	//这个滚动步骤暂时还没有找到更好的方法
	//这个设置最多能获取20页（并且浏览器视图不能超过20000长度）
	for j := 0; j < 20; j++ {
		err := mouse.Scroll(20000, 20000, j)
		if err != nil {
			break
		}
	}
	result := page.MustElement(`#J_domain`)

	pList := result.MustElements(`p`)

	for j := 0; j < len(pList); j++ {
		reverseDomain := structs.Domains{}
		a, er := pList[j].Element(`a`)
		if er != nil {
			continue
		}

		resultIP, e := a.Text()
		if e != nil {
			return OpenBrowser(remoteAddr, ua, domains, arg)
		}
		reverseDomain.Name = resultIP
		reverseDomain.IsVhost = false
		reverseDomain.CanResolve = false
		reverseDomain.Method = "reverse"

		//给结果列表添加元素
		domains = append(domains, reverseDomain)

	}
	//设置一个睡眠，避免ipchaxun.com封IP
	time.Sleep(1 * time.Second)
	return domains

}

type webScanResult struct {
	Domain string
	title  string
}

// WebScanReverseIP API接口反查IP
func WebScanReverseIP(responseList []*retryable.Response, httpHead string) []structs.Domains {
	var res []webScanResult
	var result = make([]structs.Domains, 0)
	httpBase := "http://api.webscan.cc/?action=query&ip="
	httpsBase := "https://api.webscan.cc/?action=query&ip="
	var url, remoteAddr string

	for _, response := range responseList {
		remoteAddr = response.NetConn.Conn.RemoteAddr().String()
		ip := strings.Split(remoteAddr, ":")[0]

		if strings.Contains(httpHead, "http://") {
			url = httpBase + ip
		}
		if strings.Contains(httpHead, "https://") {
			url = httpsBase + ip
		}

		response, err := http.Get(url)
		if err != nil {
			return WebScanReverseIP(responseList, httpHead)
		}

		all, err2 := ioutil.ReadAll(response.Body)
		if err2 != nil {
			return WebScanReverseIP(responseList, httpHead)
		}
		err3 := json.Unmarshal(all, &res)
		if err3 != nil {
			return WebScanReverseIP(responseList, httpHead)
		}

		var reverseDomain = structs.Domains{}
		for _, results := range res {
			reverseDomain.Name = results.Domain
			reverseDomain.IsVhost = false
			reverseDomain.CanResolve = false
			reverseDomain.Method = "reverse"
			result = append(result, reverseDomain)
		}
		break
	}

	return result
}
