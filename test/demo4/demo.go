package main

import (
	"encoding/base64"
	"fmt"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net/http"
)

const FofaEmail = "saccount2077@protonmail.com"
const FofaKey = "ea98e38b35ab4f91189d75b724dfa12f"

const ShodanKey = "JgF8iUdjxdODTma08wfw2SySkJiGLBmK"

func main() {
	FofaSearch("ip=118.112.227.0/24")

	//ShodanSearch("118.112.227.142")
}

// FofaSearch 从fofa搜索全部信息
// 按照fofa搜索语法进行请求，支持domain,host,ip,header,body,title，运算符支持== = != =~，通常情况下host就是既可以搜domain，也可以搜ip
func FofaSearch(domain string) {
	domainBase64 := base64.StdEncoding.EncodeToString([]byte(domain))
	//fmt.Println(domainBase64)
	base := fmt.Sprintf("https://fofa.so/api/v1/search/all?email=%s&key=%s&task_id=1&qbase64=%s",
		FofaEmail,
		FofaKey, domainBase64)
	response, err := http.Get(base)
	if err != nil {
		log.Fatal(err)
	}

	readAll, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(readAll))
	newJson, err := simplejson.NewJson(readAll)
	if err != nil {
		log.Fatalf("fofa转json失败:%v",err)
	}
	//e:=newJson.Get("error")
	//r:=newJson.Get("results")
	//fmt.Println(e)
	//fmt.Println(r)
	fmt.Println(newJson)

}

// ShodanSearch 从shodan搜索全部信息
func ShodanSearch(domain string) {
	base := fmt.Sprintf("https://api.shodan.io/shodan/host/%s?key=%s", domain, ShodanKey)
	response, err := http.Get(base)
	if err != nil {
		log.Fatal(err)
	}
	readAll, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(readAll))
}
