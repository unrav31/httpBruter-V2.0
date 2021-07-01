package vhost

import (
	"crypto/tls"
	"fmt"
	"github.com/hbakhtiyor/strsim"
	"github.com/rs/xid"
	"httpBruter/pkg/randHeader"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// IsVirtualHost 发包两次，探测是否使用vhost
func IsVirtualHost(url string) bool {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //忽略证书错误
	}

	client := &http.Client{Transport: tr, Timeout: 10 * time.Second}

	request, e1 := http.NewRequest("GET", url, nil)
	if e1 != nil {
		return false
	}

	request.Header.Add("User-Agent", randHeader.RandHeader())
	request.Header.Add("Referer", "https://www.google.com/")

	defer client.CloseIdleConnections() //函数结束后关闭连接

	response1, e := client.Do(request)
	if e != nil {
		return false
	}
	request.Host = fmt.Sprintf("%s.%s", xid.New().String(), request.Host)

	response2, e2 := client.Do(request) //如果请求失败直接返回false

	if e2 != nil {
		return false
	}

	if response1.StatusCode != response2.StatusCode { //判断响应码是否相等
		return true
	}

	body1, ee1 := ioutil.ReadAll(response1.Body)
	if ee1 != nil {
		return false
	}
	body2, ee2 := ioutil.ReadAll(response2.Body)
	if ee2 != nil {
		return false
	}

	if len(body1) != len(body2) {
		return true
	}

	word1 := len(strings.Split(string(body1), " ")) //判断单词个数是否相等
	word2 := len(strings.Split(string(body2), " "))

	if word1 != word2 {
		return true
	}

	line1 := len(strings.Split(string(body1), "\n")) //判断行数是否相等
	line2 := len(strings.Split(string(body2), "\n"))

	if line1 != line2 {
		return true
	}

	if strsim.Compare(string(body1), string(body2))*100 <= 100 { //相似度
		return true
	}
	return false
}
