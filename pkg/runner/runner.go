package runner

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
	"httpBruter/pkg/database"
	domains "httpBruter/pkg/domain"
	"httpBruter/pkg/ip"
	"httpBruter/pkg/match"
	"httpBruter/pkg/options"
	"httpBruter/pkg/ports"
	"httpBruter/pkg/randHeader"
	"httpBruter/pkg/retryable"
	"httpBruter/pkg/services"
	"httpBruter/pkg/structs"
	"net/http"
	"strings"
	"time"
)

func Runner(url string, arg *options.Args) {
	//创建一个响应结构体列表，里面存储DefaultURL和Fallback返回的响应内容
	responseList := make([]*retryable.Response, 0)
	data := &structs.Database{}

	var params = &retryable.Params{

		RetryTimes: arg.Retries,
		Timeout:    15,
		Cookie: &http.Cookie{
			Name:    "rememberMe",
			Value:   "1",
			Expires: time.Now().Add(1 * time.Second),
		},
		Redirects:       arg.Redirects,
		UserAgent:       randHeader.RandHeader(),
		Refer:           "https://www.google.com/hk",
		HeadlessTimeout: 60,
	}
	//构建url
	params.MakeURL(url, arg)
	//默认的https请求
	responseList = params.Request(params.DefaultURL, responseList, arg)
	//nofallback请求
	if params.FallbackURL != "" {
		responseList = append(params.Request(params.FallbackURL, responseList, arg)[:])
	}

	if len(responseList) == 0 {
		return
	}
	data.IP.IpStr = ip.IpStr(responseList)

	//开启或关闭cdn探测
	data.IP.IsReal = func(c bool) bool {
		if c {
			return ip.IsReal(params.RAW)
		} else {
			return false
		}

	}(arg.Cdn)

	//记录是哪个domain进行的ip反查
	data.IP.RealDomain = ip.RealDomain(params.RAW)

	//读取响应TEXT并解码到Response结构体里，之后要用直接去取就可以了.同时将数据库的结构体赋值
	for _, response := range responseList {
		service := structs.Services{}
		response.ToText()
		remoteAddr := services.Addr(response)

		//headless请求，获取title，并将截图转为base64编码
		if arg.Headless {
			params.Chromedp(arg, params.DefaultURL, response)
		}

		//services各成员属性
		service.Name = services.Name()
		service.Port = services.Port(remoteAddr)
		service.Addr = response.TrueURL
		service.RequestMethod = services.RequestMethod(params.RAW)
		service.Domain = services.Domain(response)
		service.Https = services.Https(response.ResponseList)
		service.CanAccess = services.CanAccess(response)

		//公共部分
		service.Title = services.Title(response.ResponseText)
		service.Method = services.Method()
		service.StatusCode = fmt.Sprintf("%d", services.StatusCode(response.ResponseList))
		service.ContentLength = fmt.Sprintf("%d", services.ContentLength(response.ResponseText))
		service.ContentType = services.ContentType(response.ResponseList)
		service.Location = services.Location(response.ResponseList)
		service.Server = services.Server(response.ResponseList)
		service.ResponseTime = services.ResponseTime(response)
		//开启或关闭finger探测
		service.Finger = func(f bool) []string {
			if f {
				return append(service.Finger, services.Finger(arg.FingerContent, params.DefaultURL, response.ResponseList))
			} else {
				return []string{}
			}
		}(arg.Finger)
		service.Screenshot = response.ScreenshotBase

		//添加到databse
		data.Services = append(data.Services, service)

	}
	//domain部分
	data.Domains = domains.DomainResults(responseList, params, arg)

	//ports部分
	data.Ports = ports.PortsResults(responseList)

	//使用正则进行匹配
	if arg.Match != "" {
		data = match.MatchTitle(data, arg.Match)
		//如果没匹配到就直接返回，进行下一个url的请求
		if data == nil {
			return
		} else {
			if arg.Clistats != nil {
				//统计一条匹配的结果
				arg.Clistats.IncrementCounter("Match", 1)
			}
		}
	}

	//入库
	if arg.Database {

		database.InsertOne(data, arg)
		if arg.Clistats != nil {
			arg.Clistats.IncrementCounter("Database", 1)
		}
	}

	//写txt文件
	if arg.OT != "" {
		data.WriteToTxt(arg)
	}

	//写json文件
	if arg.OJ != "" {
		data.WriteToJson(arg)
	}
	//控制台输出
	for _, output := range data.Services {
		_, _ = fmt.Fprintf(color.Output, "[ %s | %s | %s | %s | %s ] \n",
			runewidth.FillRight(output.Addr, 30),
			color.CyanString("%s", runewidth.FillRight(output.Title, 40)),
			color.GreenString("%s", runewidth.FillRight(strings.Join(output.Finger, ""), 15)),
			color.YellowString("%s", runewidth.FillRight(output.StatusCode, 10)),
			color.CyanString("%s", runewidth.FillRight(output.ContentLength, 10)),
		)
	}
}
