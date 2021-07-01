package parseContent

import (
	"regexp"
	"strings"
)

const title = `(?im)<\s*title.*>(.*?)<\s*/\s*title>|(?im)<\s*TITLE.*>(.*?)<\s*/\s*TITLE>`
const h1 = `<h1>(.*?)</h1>|<H1>(.*?)</H1>`

var badStatus = []int{404, 503, 500}

// ParseTitle 匹配title，传入已转换编码的bytes，返回title的字符串，匹配失败返回空字符串
func ParseTitle(readbytes []byte) string {
	var titlestr string
	Reg := regexp.MustCompile(title)
	titlestr = Reg.FindString(string(readbytes))
	if len(titlestr) != 0 {
		titleBeginIdx := strings.Index(titlestr, ">")
		titleEndIdx := strings.Index(titlestr, "</")
		subtitle := titlestr[titleBeginIdx+1 : titleEndIdx]
		return subtitle
	}
	return ""
}

// ParseH1 匹配h1标题，仅限503、404、500这类页面进行匹配，传入已解码的bytes，返回h1字符串，匹配失败返回空
func ParseH1(readbytes []byte) string {
	var h1str string
	Reg := regexp.MustCompile(h1)
	h1str = Reg.FindString(string(readbytes))
	if len(h1str) != 0 {
		h1BeginIdx := strings.Index(h1str, ">")
		h1EndIdx := strings.Index(h1str, "</")
		subh1 := h1str[h1BeginIdx+1 : h1EndIdx]
		return subh1
	}
	return ""
}

// CheckTitleNull  判断title是否为空
func CheckTitleNull(title string) bool {
	if title == "" || title == " " || title == "   " {
		return true
	} else {
		return false
	}
}

// CheckStatusCode 判断响应码,503、404、500，并且length < 200，如果成立表示不被加入到headless列表，返回false。反之加入，并返回true
func CheckStatusCode(status int, contentlen int) bool {
	var flag = true
	for i := 0; i < len(badStatus); i++ {
		if status == badStatus[i] && contentlen < 200 {
			flag = false
			return flag
		} else {
			continue
		}
	}
	return flag
}
