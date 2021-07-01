package domain

import (
	"httpBruter/pkg/options"
	"httpBruter/pkg/retryable"
	"httpBruter/pkg/structs"
	"httpBruter/pkg/vhost"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// DomainResults domain结果整合
func DomainResults(responseList []*retryable.Response, params *retryable.Params, args *options.Args) []structs.Domains {
	domainResults := make([]structs.Domains, 0)

	for _, res := range responseList {

		var domains = structs.Domains{}

		domains.Name = Name(res.ResponseList)
		//如果检测到domain的name不是域名，则跳过
		reg := regexp.MustCompile(`[a-zA-Z]`)
		if !reg.MatchString(domains.Name) {
			continue
		}

		//开启或关闭vhost探测
		domains.IsVhost = func(v bool) bool {
			if v {
				return IsVhost(params.DefaultURL)
			} else {
				return false
			}

		}(args.Vhost)
		domains.CanResolve = CanResolve(params.RAW)
		domains.Method = Method(domains.CanResolve)
		domainResults = append(domainResults, domains)
	}
	reverseDomainList := make([]structs.Domains, 0)

	//如果开启了reverseIP选项
	if args.ReverseIP {
		if args.ReverseIPOpt == "ipchaxun" {
			reverseDomainList = ReverseIP(responseList, params.UserAgent, args)
		} else if args.ReverseIPOpt == "webscan" {
			reverseDomainList = WebScanReverseIP(responseList, params.DefaultURL)
		} else if args.ReverseIPOpt == "all" {
			mapList := make(map[string]structs.Domains)

			reverseDomainList1 := ReverseIP(responseList, params.UserAgent, args)
			reverseDomainList1 = append(reverseDomainList1, WebScanReverseIP(responseList, params.DefaultURL)...)
			//排除重复元素
			for i := 0; i < len(reverseDomainList1); i++ {
				mapList[reverseDomainList1[i].Name] = reverseDomainList1[i]
			}
			//结果加入最后返回的列表中
			for _, v := range mapList {
				reverseDomainList = append(reverseDomainList, v)
			}

		} else {
			log.Fatal("[x] 反查IP的接口不存在，请选择'go-rod'或'api'")
		}
	}

	return append(domainResults, reverseDomainList...)
}

func Name(response []*http.Response) string {
	name := ""
	for _, res := range response {
		name = res.Header.Get("Host")
		URL := res.Request.URL.String()
		reg := regexp.MustCompile(`[a-zA-Z]+`)

		if name == "" && reg.MatchString(URL) {
			name = strings.Replace(URL, "https://", "", -1)
			name = strings.Replace(name, "http://", "", -1)
			if strings.Contains(name, ":") {
				name = strings.Split(name, ":")[0]
			}
		}
		break
	}
	return strings.Replace(name, "/", "", -1)
}

func IsVhost(url string) bool {
	return vhost.IsVirtualHost(url)
}

func CanResolve(RAW string) bool {
	RAW = strings.Replace(RAW, "http://", "", -1)
	RAW = strings.Replace(RAW, "https://", "", -1)
	reg := regexp.MustCompile(`[a-zA-Z]+`)
	if reg.MatchString(RAW) {
		return true
	} else {
		return false
	}
}

func Method(canResolve bool) string {
	if canResolve {
		return "direct"
	} else {
		return "reverse"
	}

}
