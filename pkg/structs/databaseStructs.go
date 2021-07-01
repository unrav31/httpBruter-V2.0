package structs

import (
	"encoding/json"
	"fmt"
	"httpBruter/pkg/options"
	"log"
	"strings"
)

type ip struct {
	IpStr      string `json:"ip_str"`
	IsReal     bool   `json:"isReal"`
	RealDomain string `json:"realDomain"`
}

type Domains struct {
	Name       string `json:"name"`       //response的host值
	IsVhost    bool   `json:"isVhost"`    //调用vhost来赋值
	CanResolve bool   `json:"canResolve"` //只要是域名，都为true
	Method     string `json:"method"`     //域名为direct，ip为reverse
}

type Ports struct {
	Name    string `json:"name"`    //web
	Value   string `json:"value"`   //端口号，在remoteAddr里取
	Method  string `json:"method"`  //brute
	State   string `json:"state"`   //根据响应来判断是否open或close
	Product string `json:"product"` //在响应头中获取服务器类型
	Version string `json:"version"` //响应头中获取服务器版本
	Iwp     string `json:"iwp"`     //就是响应头中的remoteAddr
}

type Services struct {
	Name          string   `json:"name"` //web
	Port          string   `json:"port"` //remoteAddr里面的端口号
	Addr          string   `json:"addr"` //请求的url
	RequestMethod string   `json:"requestMethod"`
	Domain        string   `json:"domain"`
	Https         bool     `json:"https"`
	CanAccess     bool     `json:"canAccess"`
	Title         string   `json:"title"`
	Method        string   `json:"Method"`
	StatusCode    string   `json:"statusCode"`
	ContentLength string   `json:"contentLength"`
	ContentType   string   `json:"contentType"`
	Location      []string `json:"location"`
	Server        string   `json:"server"`
	ResponseTime  string   `json:"responseTime"`
	Finger        []string `json:"finger"`
	Screenshot    string   `json:"screenshot"`
}

// Database 按照Notion所给出的json格式构造结构体
type Database struct {
	IP       ip         `json:"ip"`
	Domains  []Domains  `json:"domains"`
	Ports    []Ports    `json:"ports"`
	Services []Services `json:"services"`
}

func (d *Database) WriteToTxt(args *options.Args) {
	builder := strings.Builder{}

	builder.WriteString("[ " + d.Services[0].Addr + " | " + d.Services[0].Title + " | ")
	if args.Finger {
		builder.WriteString(strings.Join(d.Services[0].Finger, "") + " | ")
	}
	builder.WriteString(d.Services[0].StatusCode + " | ")
	builder.WriteString(d.Services[0].ContentLength + " ]")
	_, err := args.WriteResults.WriteString(builder.String() + "\n")
	if err != nil {
		fmt.Println("[x] 写入文件失败", builder.String())
		log.Fatal(err)
	}
}

type jsonData struct {
	Url           string   `json:"url"`
	Location      []string `json:"location"`
	Title         string   `json:"title"`
	WebServer     string   `json:"webServer"`
	ContentType   string   `json:"contentType"`
	Method        string   `json:"method"`
	ContentLength string   `json:"contentLength"`
	StatusCode    string   `json:"statusCode"`
	Vhost         bool     `json:"vhost"`
	Cdn           bool     `json:"cdn"`
	ResponseTime  string   `json:"responseTime"`
	Finger        []string `json:"finger"`
}

func (d *Database) WriteToJson(args *options.Args) {
	data := jsonData{}
	data.Url = d.Services[0].Addr
	data.Location = d.Services[0].Location
	data.Title = d.Services[0].Title
	data.WebServer = d.Services[0].Server
	data.ContentType = d.Services[0].ContentType
	data.Method = d.Services[0].Method
	data.ContentLength = d.Services[0].ContentLength
	data.StatusCode = d.Services[0].StatusCode
	data.Vhost = func(dl []Domains) bool {
		if len(dl) == 0 {
			return false
		} else {
			return dl[0].IsVhost
		}
	}(d.Domains)
	data.Cdn = d.IP.IsReal
	data.ResponseTime = d.Services[0].ResponseTime
	data.Finger = d.Services[0].Finger

	marshal, err := json.Marshal(data)
	if err != nil {
		fmt.Println("[x] 写入文件时转化json错误", string(marshal))
		log.Fatal(err)
	}

	_, err2 := args.WriteJsonResults.WriteString(string(marshal) + "\n")
	if err2 != nil {
		fmt.Println("[x] 写入json文件时错误", err)
		log.Fatal(err)
	}
}
