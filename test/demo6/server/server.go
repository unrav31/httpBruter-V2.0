package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp","127.0.0.1:9000")
	if err != nil {
		log.Fatal(err)
	}
	for {
		accept, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go testServer(accept)
	}
}

func testServer(connect net.Conn)  {
	defer func(connect net.Conn) {
		err := connect.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(connect)

	for{
		reader:=bufio.NewReader(connect)
		var buf [4096]byte
		n, err := reader.Read(buf[:])
		if err != nil {
			log.Fatal(err)
		}
		recv:=string(buf[:n])
		fmt.Printf("收到的数据：%v\n", recv)
		// 将接受到的数据返回给客户端
		_, err = connect.Write([]byte("ok"))
		if err != nil {
			log.Fatal(err)
		}
	}
}