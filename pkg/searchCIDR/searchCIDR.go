package searchCIDR

import (
	"github.com/beevik/etree"
	"httpBruter/pkg/randHeader"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ReverseCIDR 反查源IP的C段主逻辑，传入全局变量globalCIDRMap，如果true，则reverseIP，如果false，则不去reverse了
func ReverseCIDR(remoteAddr string, globalCIDRMap map[string][]string) bool {
	//源ip的C段
	remoteIP := strings.Split(remoteAddr, ":")[0]
	remoteCIDRLs := strings.Split(remoteIP, ".")
	remoteCIDR := remoteCIDRLs[0] + "." + remoteCIDRLs[1] + "." + remoteCIDRLs[2] + ".0/24"

	if len(globalCIDRMap[remoteCIDR]) != 0 {
		return SearchIP(remoteIP, globalCIDRMap[remoteCIDR])
	} else {
		globalCIDRMap = SearchCIDR(remoteIP, globalCIDRMap)
		return SearchIP(remoteIP, globalCIDRMap[remoteCIDR])
	}
}

// SearchIP 在接口返回的结果列表中查找IP，查找到了返回true，没查找到返回false
func SearchIP(remoteIP string, cidrResults []string) bool {
	for i := 0; i < len(cidrResults); i++ {
		if remoteIP == cidrResults[i] {
			return true
		} else {
			continue
		}
	}
	return false
}

// SearchCIDR 获取C段信息，将map后的结果存入全局变量中
func SearchCIDR(rawIP string, globalCIDRMap map[string][]string) map[string][]string {

	//拼接C段作为map的key
	ipList := strings.Split(rawIP, ".")
	key := ipList[0] + "." + ipList[1] + "." + ipList[2] + ".0/24"
	url := "https://chapangzhan.com/"

	client := &http.Client{Timeout: 10 * time.Second}
	request, err := http.NewRequest("GET", url+rawIP, nil)
	if err != nil {
		return SearchCIDR(rawIP, globalCIDRMap)
	}

	request.Header.Add("User-Agent", randHeader.RandHeader())
	response, er := client.Do(request)
	if er != nil {
		return SearchCIDR(rawIP, globalCIDRMap)
	}
	line, _ := ioutil.ReadAll(response.Body)

	Reg := regexp.MustCompile(`<tbody>[.\s\S]+</tbody>`)
	regResults := Reg.FindAllString(string(line), -1)

	if len(regResults) == 0 {
		log.Fatal("[x] 未匹配到网页响应")
	}
	t := regResults[0]

	// 初始化根节点
	doc := etree.NewDocument()
	err1 := doc.ReadFromString(t)
	if err1 != nil {
		log.Fatal("[x] etree解析失败")
	}
	root := doc.SelectElement("tbody")

	v := root.FindElements("./tr")
	if len(v) == 0 {
		return SearchCIDR(rawIP, globalCIDRMap)
	}

	reverseIPList := make([]string, 0)

	for i := 0; i < len(v); i++ {

		//IP
		a := v[i].FindElement("./td/a")
		if a == nil {
			continue
		}
		aa := a.Text()

		//count
		k := v[i].FindElement("./td/span[0]")
		if k == nil {
			continue
		}

		kk := k.Text()
		count, _ := strconv.Atoi(kk)

		//若count为0，表示这个IP没有结果，跳过，否则加入反查ip的列表
		if count == 0 {
			continue
		} else {
			reverseIPList = append(reverseIPList, strings.Replace(aa, " ", "", -1))
		}
	}
	globalCIDRMap[key] = reverseIPList
	return globalCIDRMap
}
