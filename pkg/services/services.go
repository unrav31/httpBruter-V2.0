package services

import (
	"fmt"
	"httpBruter/pkg/finger"
	"httpBruter/pkg/retryable"
	"net/http"
	"regexp"
	"strings"
)

const (
	//titlePattern  = `(?im)<\s*title>([\s\S]+)<\s*/\s*title>|(?im)<\s*TITLE>([\s\S]+)<\s*/\s*TITLE>`
	titlePattern  = `(?im)<title>(.*?)</title>|(?im)<TITLE>(.*?)</TITLE>`
	h1            = `<h1>([\s\S]+)</h1>|<H1>([\s\S]+)</H1>`
	domainPattern = `[a-zA-Z]+`
	ipPattern     = `.*?(\d+\.\d+\.\d+\.\d+).*?`
)

func Name() string {
	return "web"
}

func Port(remoteAddr string) (port string) {
	port = strings.Split(remoteAddr, ":")[1]
	return
}

// Addr remoteAddr,远程IP地址(带端口号)
func Addr(response *retryable.Response) (remoteAddr string) {
	remoteAddr = response.NetConn.Conn.RemoteAddr().String()
	return
}

// Location 获取重定向地址
func Location(responseList []*http.Response) []string {
	locationList := make([]string, 0)

	for i := 0; i < len(responseList); i++ {
		location, err := responseList[i].Location()
		if err != nil {
			continue
		}
		locationList = append(locationList, location.String())
	}
	return locationList

}

func RequestMethod(RAW string) string {
	raw := strings.Replace(RAW, "https://", "", -1)
	raw = strings.Replace(raw, "http://", "", -1)

	regDomain := regexp.MustCompile(domainPattern)
	regIP := regexp.MustCompile(ipPattern)

	if strings.Contains(raw, ":80") || strings.Contains(raw, ":443") {
		return "brute"
	} else if strings.Contains(raw, ":") && regIP.MatchString(raw) {
		return "iwp"
	} else if strings.Contains(raw, ":") && regDomain.MatchString(raw) {
		return "dwp"
	} else {
		return "brute"
	}
}

func Domain(response *retryable.Response) string {
	domain := ""
	for _, res := range response.ResponseList {
		domain = res.Header.Get("Host")
		URL := res.Request.URL.String()
		reg := regexp.MustCompile(`[a-zA-Z]+`)

		if domain == "" && reg.MatchString(URL) {
			domain = strings.Replace(URL, "https://", "", -1)
			domain = strings.Replace(domain, "http://", "", -1)
			if strings.Contains(domain, ":") {
				domain = strings.Split(domain, ":")[0]
			}
		}
		break
	}
	//如果匹配是IP，那么就返回空
	reg := regexp.MustCompile(`[a-zA-Z]`)
	if !reg.MatchString(domain) {
		return ""
	}
	return domain
}

func Https(responseList []*http.Response) bool {
	for _, response := range responseList {
		if strings.Contains(response.Request.URL.String(), "https://") {
			return true
		} else {
			continue
		}
	}
	return false
}

func CanAccess(response *retryable.Response) bool {
	return response.HttpCanAccess
}

func Title(responseText []string) string {
	response := ""
	for _, res := range responseText {
		response = res
		break
	}
	responseBytes := []byte(response)
	var title string
	Reg := regexp.MustCompile(titlePattern)
	title = Reg.FindString(string(responseBytes))
	if len(title) != 0 && title != "" {
		titleBeginIdx := strings.Index(title, ">")
		titleEndIdx := strings.Index(title, "</")
		titleStrip := strings.Replace(title[titleBeginIdx+1:titleEndIdx], "\r", "", -1)
		titleStrip = strings.Replace(titleStrip, "\n", "", -1)
		titleStrip = strings.Replace(titleStrip, "\t", "", -1)
		return titleStrip

	}

	RegH1 := regexp.MustCompile(h1)
	title = RegH1.FindString(string(responseBytes))
	if len(title) != 0 && title != "" {
		h1BeginIdx := strings.Index(title, ">")
		h1EndIdx := strings.Index(title, "</")
		titleStrip := strings.Replace(title[h1BeginIdx+1:h1EndIdx], "\r", "", -1)
		titleStrip = strings.Replace(titleStrip, "\n", "", -1)
		titleStrip = strings.Replace(titleStrip, "\t", "", -1)
		return titleStrip
	} else {
		return ""
	}
}

func Method() string {
	return "GET"
}

func StatusCode(response []*http.Response) (status int) {
	for _, res := range response {
		status = res.StatusCode
		break
	}
	return status
}

func ContentLength(responseList []string) int {
	var contentLength int
	for _, res := range responseList {
		contentLength = len([]byte(res))
		break
	}
	return contentLength
}

func ContentType(response []*http.Response) string {
	contentType := ""
	for _, res := range response {
		contentType = res.Header.Get("Content-Type")
		break
	}
	return contentType
}

func Server(response []*http.Response) string {
	var server string
	for _, res := range response {
		server = res.Header.Get("Server")
		break
	}
	return server
}

func ResponseTime(response *retryable.Response) string {
	return fmt.Sprintf("%d ms", response.ResponseTime)
}

func Finger(fingerData []finger.Content, url string, responseList []*http.Response) string {

	return finger.NormalFinger(fingerData, url, responseList)
}
