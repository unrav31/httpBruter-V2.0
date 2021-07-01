package ip

import (
	"httpBruter/pkg/cdn"
	"httpBruter/pkg/retryable"
	"regexp"
	"strings"
)

// IpStr 传入带冒号的remoteAddr，返回ip_str
func IpStr(responseList []*retryable.Response) string {
	remoteAddr := ""
	for _, response := range responseList {
		remoteAddr = response.NetConn.Conn.RemoteAddr().String()
	}
	return strings.Split(remoteAddr, ":")[0]
}

// IsReal 传入domain，不能传ip
func IsReal(domain string) bool {
	return cdn.SuperPing(domain)
}

// RealDomain 记录是哪个domain使用的cdn探测
func RealDomain(RAW string) string {
	//匹配domain,如果不是domain，则返回空字符串
	RAW = strings.Replace(RAW, "http://", "", -1)
	RAW = strings.Replace(RAW, "https://", "", -1)
	reg := regexp.MustCompile(`[a-zA-Z]`)
	if reg.MatchString(RAW) {
		return RAW
	} else {
		return ""
	}
}
