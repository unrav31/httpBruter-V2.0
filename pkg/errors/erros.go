package errors

import (
	"fmt"
	"github.com/fatih/color"
	"log"
	"os"
	"strings"
)

func HandleRequestErrors(url string, err error, debug bool) {
	if !debug {
		return
	}

	if strings.Contains(err.Error(), "EOF") {
		fmt.Printf("%s ", url)
		color.Red("[x] 连接错误(重置)")
	} else if strings.Contains(err.Error(), "no such host") {
		fmt.Printf("%s ", url)
		color.Red("[x] 连接错误(找不到服务器IP地址)")
	} else if strings.Contains(err.Error(), "closed by the remote host") {
		fmt.Printf("%s ", url)
		color.Red("[x] 连接错误(重置)")
	} else if strings.Contains(err.Error(), "context deadline exceeded") {
		fmt.Printf("%s ", url)
		color.Red("[x] 连接错误(重置)") //也是超时
	} else if strings.Contains(err.Error(), "internal error") {
		fmt.Printf("%s ", url)
		color.Red("[x] 连接错误(发送的响应无效)")
	} else if strings.Contains(err.Error(), "client") {
		fmt.Printf("%s ", url)
		color.Red("[x] 连接错误(发送的响应无效)")
	} else if strings.Contains(err.Error(), "connection reset") {
		fmt.Printf("%s ", url)
		color.Red("[x] 连接错误(重置)")
	} else if strings.Contains(err.Error(), "refused") {
		fmt.Printf("%s ", url)
		color.Red("[x] 连接错误(拒绝)")
	} else if strings.Contains(err.Error(), "handshake failure") {
		fmt.Printf("%s ", url)
		color.Red("[x] 连接错误(无法提供安全连接)")
	} else if strings.Contains(err.Error(), "CONNECTION_RESET") {
		fmt.Printf("%s ", url)
		color.Red("[x] 连接错误(重置)")
	} else if strings.Contains(err.Error(), "deadline exceeded") {
		fmt.Printf("%s ", url)
		color.Red("[x] 连接错误(超时)")
	} else if strings.Contains(err.Error(), "EMPTY_RESPONSE") {
		fmt.Printf("%s ", url)
		color.Red("[x] 连接错误(未发送任何数据)")

	} else if strings.Contains(err.Error(), "SSL_PROTOCOL_ERROR") {
		fmt.Printf("%s ", url)
		color.Red("[x] 连接错误(响应无效)")

	} else if strings.Contains(err.Error(), "NOT_RESOLVED") {
		fmt.Printf("%s ", url)
		color.Red("[x] 连接错误(找不到服务器IP地址)")

	} else if strings.Contains(err.Error(), "CONNECTION_CLOSED") {
		fmt.Printf("%s ", url)
		color.Red("[x] 连接错误(意外终止了连接)")

	} else if strings.Contains(err.Error(), "UNSAFE_PORT") {
		fmt.Printf("%s ", url)
		color.Red("[x] 连接错误(无法连接)")

	} else if strings.Contains(err.Error(), "ABORTED") {
		fmt.Printf("%s ", url)
		color.Red("[x] 连接错误(错误页面)")

	} else if strings.Contains(err.Error(), "TIMED_OUT") {
		fmt.Printf("%s ", url)
		color.Red("[x] 连接错误(TIMED_OUT)")
	} else if strings.Contains(err.Error(), "file does not exist") || strings.Contains(err.Error(), "directory") {
		color.Red("[x] 指定chrome程序无效")
		os.Exit(0)
	} else if strings.Contains(err.Error(), "application") || strings.Contains(err.Error(), "permission") {
		color.Red("[x] 指定chrome程序无效")
		os.Exit(0)
	} else {
		color.Red("[x] 其他错误，请排查", url, err.Error())
		log.Fatal(err)
	}

}
