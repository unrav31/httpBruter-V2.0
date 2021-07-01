package parseCIDR

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

const (
	URLPattern = `[a-zA-Z]+`
	IPPattern  = `.*?(\d+\.\d+\.\d+\.\d+).*?`
)

// ParseCIDR 读取文件后的第一步处理，解析C段，单个IP或URL直接加入结果列表，不对URL做任何处理，此结果用于RetryableRequest中的RAW
// 此结果不能用于请求，还需要调用MakeURL。
func ParseCIDR(readList []string) (result []string) {

	for i := 0; i < len(readList); i++ {
		cache := strip(readList[i])

		reg := regexp.MustCompile(URLPattern)
		regIP := regexp.MustCompile(IPPattern)

		if reg.MatchString(cache) {
			result = append(result, readList[i])

		} else if strings.Contains(cache, ":") {
			result = append(result, readList[i])

		} else if regIP.MatchString(cache) && strings.Contains(cache, "/") {
			slashResultList := slash(cache)
			for j := 0; j < len(slashResultList); j++ {
				result = append(result, slashResultList[j])
			}

		} else if regIP.MatchString(cache) && strings.Contains(cache, "-") {
			lineResultList := line(cache)
			for j := 0; j < len(lineResultList); j++ {
				result = append(result, lineResultList[j])
			}

		} else if regIP.MatchString(cache) {
			//单个IP类型的URL
			result = append(result, readList[i])

		} else {
			log.Fatalf("[x] IP解析 %s 无法解析的类型", readList[i])
		}

	}
	return
}

//去除http和https字段，只保留域名或者IP，如果字符串最后是'/'，则去除斜杠
func strip(old string) (new string) {

	new = strings.Replace(old, "http://", "", -1)
	new = strings.Replace(new, "https://", "", -1)

	end := new[len(new)-1]

	if end == '/' {
		new = new[:len(new)-2]
	}

	return
}

//处理IP段列表，默认只能处理24及大于24，小于24的情况会报错
func slash(cidrIP string) (result []string) {

	slashList := strings.Split(cidrIP, "/")

	splitList := strings.Split(slashList[0], ".")
	domain, _ := strconv.Atoi(slashList[1])

	if domain < 24 {
		log.Fatalf("[x] IP解析 %s 不支持大于255的地址", cidrIP)
	}

	sar := uint32(0xffffffff) >> uint32(domain)
	cidrBegin, _ := strconv.Atoi(splitList[3])

	for i := cidrBegin; i < int(sar)+1; i++ {

		newIP := splitList[0] + "." + splitList[1] + "." + splitList[2] + "." + strconv.Itoa(i)
		result = append(result, newIP)
	}
	return
}

//处理以横线分割的段列表，有两种情况："192.168.1.0-255" 和 "192.168.1.0-192.168.1.255"
func line(lineIP string) (result []string) {
	lineList := strings.Split(lineIP, "-")

	//区别 "192.168.1.0-192.168.1.255"
	if len(lineList[1]) > 3 {
		firstIPSplitList := strings.Split(lineList[0], ".")
		sencondIPSplitList := strings.Split(lineList[1], ".")

		cidrBegin, _ := strconv.Atoi(firstIPSplitList[3])
		cidrEnd, _ := strconv.Atoi(sencondIPSplitList[3])

		for i := cidrBegin; i < cidrEnd+1; i++ {
			newIP := firstIPSplitList[0] + "." + firstIPSplitList[1] + "." + firstIPSplitList[2] + "." + strconv.Itoa(i)
			result = append(result, newIP)
		}

	} else {
		splitList := strings.Split(lineList[0], ".")

		cidrBegin, _ := strconv.Atoi(splitList[3])
		cidrEnd, _ := strconv.Atoi(lineList[1])

		for i := cidrBegin; i < cidrEnd+1; i++ {
			newIP := splitList[0] + "." + splitList[1] + "." + splitList[2] + "." + strconv.Itoa(i)
			result = append(result, newIP)
		}
	}
	return
}
