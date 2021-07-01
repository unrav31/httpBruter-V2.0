package finger

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/spaolacci/murmur3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type FingerContent struct {
	Fingerprint []Content `json:"fingerprint"`
}
type Content struct {
	Cms      string   `json:"cms"`
	Method   string   `json:"method"`
	Location string   `json:"location"`
	Keyword  []string `json:"keyword"`
}

// NormalFinger 获取指纹信息
func NormalFinger(fingerDataList []Content, URL string, responseList []*http.Response) string {
	allResponseInOneStr := ""

	for i := 0; i < len(responseList); i++ {
		line, _ := ioutil.ReadAll(responseList[i].Body)
		allResponseInOneStr += string(line)
	}

	result := ""

	favHash := strconv.Itoa(int(getIcon(URL)))
	//这个header应该取所有的历史响应，因为有些是在跳转前，有些是在跳转后才会有指纹
	//把所有的头部拼接起来组成一个字符串进行搜索
	allHeaderInOneStr, _ := getHeaders(responseList)

	for i := 0; i < len(fingerDataList); i++ {
		h, _ := strconv.Atoi(fingerDataList[i].Keyword[0])

		pattern := strings.Join(fingerDataList[i].Keyword, "&")
		Reg := regexp.MustCompile(pattern)
		if fingerDataList[i].Location == "body" &&
			Reg.MatchString(allResponseInOneStr) &&
			fingerDataList[i].Method == "keyword" {
			result = fingerDataList[i].Cms
			break

		}
		if fingerDataList[i].Location == "body" &&
			fingerDataList[i].Method == "faviconhash" &&
			favHash == strconv.Itoa(int(uint32(h))) {
			result = fingerDataList[i].Cms
			break
		}
		if fingerDataList[i].Location == "header" &&
			fingerDataList[i].Method == "keyword" &&
			Reg.MatchString(allHeaderInOneStr) {
			result = fingerDataList[i].Cms
			break
		}
	}
	return result
}

func OpenFinger() ([]Content, *os.File) {
	path, _ := os.Getwd()
	path = filepath.Join(path, "finger.json")

	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("[x] 错误路径 %s 指纹库读取失败", path)
	}
	var fing FingerContent
	readBytes, _ := ioutil.ReadAll(file)
	_ = json.Unmarshal(readBytes, &fing)

	fingerList := fing.Fingerprint
	return fingerList, file
}

func getIcon(url string) uint32 {
	url = url + "/favicon.ico"
	resp, err := http.Get(url)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0
	}
	encoded := base64.StdEncoding.EncodeToString(body)
	var buffer bytes.Buffer
	for i := 0; i < len(encoded); i++ {
		ch := encoded[i]
		buffer.WriteByte(ch)
		if (i+1)%76 == 0 {
			buffer.WriteByte('\n')
		}
	}
	buffer.WriteByte('\n')

	var h32 = murmur3.New32()
	_, _ = h32.Write(buffer.Bytes())
	return h32.Sum32()
}

func getHeaders(responseList []*http.Response) (string, error) {
	allHeaderInOneStr := ""

	for i := 0; i < len(responseList); i++ {

		headerContent, e := json.Marshal(responseList[i].Header)

		if e != nil {
			fmt.Printf("[!] 获取指纹时头部解析失败 %v", responseList[i].Header)
			return "", e
		} else {
			allHeaderInOneStr += string(headerContent)
		}
	}
	return allHeaderInOneStr, nil

}
