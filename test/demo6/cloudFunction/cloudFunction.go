package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tencentyun/scf-go-lib/cloudfunction"
	"github.com/tencentyun/scf-go-lib/events"
	"net"
	"os"
	"strconv"
)

type result struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	Uid  string `json:"uid"`
}

var BridgeIP = "59.110.46.110"
var BridgePort = "1234"

func hello(ctx context.Context, event events.APIGatewayRequest) {

	var b = result{}
	body := event.Body

	err := json.Unmarshal([]byte(body), &b)
	if err != nil {
		return
	}

	host := b.Host
	port := b.Port
	uid := b.Uid

	//服务器端，反向连接客户端
	server := receive(host, port)
	defer func(server net.Conn) {
		err := server.Close()
		if err != nil {
			os.Exit(-1)
		}
	}(server)
	//云函数端，连接服务器端
	bridge := send(BridgeIP, BridgePort)
	defer func(bridge net.Conn) {
		err := bridge.Close()
		if err != nil {
			os.Exit(-1)
		}
	}(bridge)
	_, err = bridge.Write([]byte(uid))

	if err != nil {
		return
	}

	serverBytes := make([]byte, 4096)
	bridgeBytes := make([]byte, 4096)

	serverChan := make(chan []byte, 1)
	bridgeChan := make(chan []byte, 1)

	for {
		select {
		//当缓冲可读，表示在服务器端接收到了客户端的输入，读出数据，并写入云函数端
		case <-serverChan:
			read, err := server.Read(serverBytes)
			if err != nil {
				return
			}
			fmt.Println("READ NUM==", read)
			fmt.Println("READ BYTES==", string(serverBytes))
			serverBytes = <-serverChan
			writeBytes, err2 := bridge.Write(serverBytes)
			if err2 != nil {
				return
			}
			fmt.Println("WRITE NUM==", writeBytes)
			fmt.Println("WRITE BYTES==", string(serverBytes))

		case <-bridgeChan:
			read, err := bridge.Read(bridgeBytes)
			if err != nil {
				return
			}
			fmt.Println("BRIDGE READ NUM==", read)
			fmt.Println("BRIDGE READ BYTES==", string(bridgeBytes))
			bridgeBytes = <-bridgeChan
			write2, err := server.Write(bridgeBytes)
			if err != nil {
				return
			}
			fmt.Println("BRIDGE WRITE NUM==", write2)
			fmt.Println("BRIDGE WRITE BYTES==", string(bridgeBytes))
		default:

		}
	}
}

//接受客户端的IP地址，并反向连接客户端
func receive(host string, port int) net.Conn {
	server, err := net.Dial("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		return nil
	}
	return server
}

//连接服务端的IP地址，并发送数据
func send(bridgeIP string, bridgePort string) net.Conn {
	bridge, err := net.Dial("tcp", bridgeIP+":"+bridgePort)
	if err != nil {
		return nil
	}
	return bridge
}

func main() {
	// Make the handler available for Remote Procedure Call by Cloud Function
	cloudfunction.Start(hello)
}
