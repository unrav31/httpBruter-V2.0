package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	for i := 0; i < 10000; i++ {

		url := "https://ipchaxun.com/59.100.46.110"
		client := http.Client{}
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36")
		request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
		//request.Header.Add("Accept-Encoding","gzip, deflate, br")
		request.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")
		request.Header.Add("Cache-Control", "max-age=0")
		request.Header.Add("Connection", "keep-alive")
		request.Header.Add("Host", "ipchaxun.com")
		request.Header.Add("sec-ch-ua", "Not;A Brand\";v=\"99\", \"Google Chrome\";v=\"91\", \"Chromium\";v=\"91\"")
		request.Header.Add("sec-ch-ua-mobile", "?0")
		request.Header.Add("Sec-Fetch-Dest", "document")
		request.Header.Add("Sec-Fetch-Mode", "navigate")
		request.Header.Add("Sec-Fetch-Site", "none")
		request.Header.Add("Sec-Fetch-User", "?1")
		request.Header.Add("Upgrade-Insecure-Requests", "1")
		res, err := client.Do(request)
		if err != nil {
			fmt.Println("eof")
			log.Fatal(err)
		}
		//all, err := ioutil.ReadAll(res.Body)
		//if err != nil {
		//	log.Fatal(err)
		//}
		defer res.Body.Close()
		//fmt.Println(string(all))
		time.Sleep(1*time.Second)
		fmt.Printf("%d--%s\n",i,res.Request.Header)
	}
}
