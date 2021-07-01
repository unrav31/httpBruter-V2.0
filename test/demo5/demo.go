package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

func main() {
	var sign = make(chan error, 1)
	var sign2 = make(chan error, 1)
	var sign3 = make(chan error, 1)
	u := "https://ipchaxun.com/59.110.46.110"
	for i := 0; i < 10000; i++ {
		select {
		//本地IP测试，一直尝试读取信号，如果读到了，就执行函数,如果请求到的err为空则停止函数
		case <-sign:
			err := requestNoProxy(u, 10)
			if err == nil {
				sign3 <- err
			}
			fmt.Println("sign")

		//socks请求，一直尝试读取信号，如果读到了，就执行函数，如果读不到就阻塞
		case <-sign2:

			err := request(u)
			if err != nil {
				fmt.Println(u, err.Error())
			}
			fmt.Println("sign2")

		//跳出循环进入下一个循环
		case <-sign3:

			fmt.Println("sign3")
			continue

		//本地IP正常请求，发送错误信号给上面两个步骤
		default:
			err := requestNoProxy(u, 0)
			fmt.Println("default")
			if err != nil {
				sign <- err
				sign2 <- err
			}
		}
	}
}

//本地IP访问，用于正常访问和check
func requestNoProxy(urls string, delay time.Duration) error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, e := http.NewRequest("GET", urls, nil)
	if e != nil {
		return e
	}

	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36")
	req.Header.Add("Host", "ipchaxun.com")

	response, err2 := client.Do(req)
	if err2 != nil {
		return err2
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(response.Body)
	time.Sleep(delay * time.Second)
	return nil
}

//socks5访问
func request(urls string) error {
	now := time.Now().UnixNano()
	proxy := "socks5://test:test@175.24.16.184:52021"
	parse, err := url.Parse(proxy)
	if err != nil {
		return err
	}
	tr := http.Transport{Proxy: http.ProxyURL(parse)}
	client := &http.Client{Transport: &tr, Timeout: 10 * time.Second}

	req, e := http.NewRequest("GET", urls, nil)
	if e != nil {
		return e
	}
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36")
	req.Header.Add("Host", "ipchaxun.com")

	response, err2 := client.Do(req)
	if err2 != nil {
		return err2
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(response.Body)

	end := time.Now().UnixNano()
	fmt.Printf("proxyTime==%d ms\n", (end-now)/1e6)

	time.Sleep(1 * time.Second)
	return nil
}
