package cdn

import (
	"encoding/json"
	"fmt"
	"httpBruter/pkg/randHeader"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type CDNStruct struct {
	Code int `json:"code"`
	Data data
	Msg  string `json:"msg"`
}
type data struct {
	Ip        string `json:"Ip"`
	Time      string `json:"Time"`
	Ttl       string `json:"Ttl"`
	IpAddress string `json:"IpAddress"`
}

// SuperPing 使用多地ping探测是否开启CDN,必须传入URL，不可以传入IP
func SuperPing(host string) bool {

	host = strings.Replace(host, "http://", "", -1)
	host = strings.Replace(host, "https://", "", -1)

	resList := make([]string, 0)
	results := firstPost(host)
	resultMap := make(map[string]string)

	ch := make(chan bool, len(results))

	for i := 0; i < len(results); i++ {
		go func(s string) {
			res := secondPost(s, host)
			resList = append(resList, res)
			ch <- true
		}(results[i][1])
	}

	for i := 0; i < len(results); i++ {
		<-ch
	}

	for j := 0; j < len(resList); j++ {
		resultMap[resList[j]] = "v"
	}
	if len(resultMap) >= 2 {
		return true
	}
	return false

}

// FirstPost post请求api逻辑函数
func firstPost(host string) [][]string {

	var params string

	node := "1,2,3,4,5,6"
	apiURL := "https://wepcc.com/"
	params = fmt.Sprintf("host=%s&node=%s", host, node)
	response := requestAPI(params, apiURL)
	line, _ := ioutil.ReadAll(response.Body)
	Reg := regexp.MustCompile(`data-id="(.*?)"`)
	result := Reg.FindAllStringSubmatch(string(line), -1)
	if len(result) < 8 {
		return firstPost(host)
	}
	result = result[:8] //选取前8个节点
	return result
}

func secondPost(result string, host string) string {
	var pingResult CDNStruct

	node2 := result
	apiURL2 := "https://wepcc.com/check-ping.html"
	params2 := fmt.Sprintf("node=%s&host=%s", node2, host)
	response2 := requestAPI(params2, apiURL2)
	line, _ := ioutil.ReadAll(response2.Body)
	e := json.Unmarshal(line, &pingResult)
	if e != nil {
		return secondPost(result, host)
	}

	return pingResult.Data.Ip
}

// RequestAPI post请求api接口，传入post参数和要post的url，返回响应
func requestAPI(params string, apiURL string) *http.Response {

	client := &http.Client{Timeout: 10 * time.Second}
	request, err := http.NewRequest("POST", apiURL, strings.NewReader(params))
	if err != nil {
		return requestAPI(params, apiURL)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	request.Header.Add("User-Agent", randHeader.RandHeader())
	request.Header.Add("referer", "https://www.google.com/hk")
	defer client.CloseIdleConnections()
	res, err := client.Do(request)
	if err != nil {
		return requestAPI(params, apiURL)
	}
	return res
}
