package decode

import (
	"bytes"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"regexp"
	"strings"
	"unicode/utf8"
)

func MainParse(responseBytes []byte) []byte {
	reg := regexp.MustCompile(`<meta.*?charset.*?>`)
	compileBytes := reg.Find(responseBytes)
	if strings.Contains(string(compileBytes), "gbk") {
		return decodeGBK(responseBytes)
	}
	if strings.Contains(string(compileBytes), "utf-8") {
		return responseBytes
	}
	if strings.Contains(string(compileBytes), "gb2312") {
		return decodeGBK(responseBytes)
	}

	if utf8.Valid(responseBytes) {
		return responseBytes
	} else {
		return decodeGBK(responseBytes)
	}
}

//GBK、GB2312解码方式
func decodeGBK(response []byte) []byte {
	gbkReader := transform.NewReader(bytes.NewReader(response), simplifiedchinese.GBK.NewDecoder())
	all, _ := ioutil.ReadAll(gbkReader)
	return all
}
