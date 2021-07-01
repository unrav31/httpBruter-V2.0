package ports

import (
	"httpBruter/pkg/retryable"
	"httpBruter/pkg/structs"
	"net/http"
	"regexp"
	"strings"
)

// PortsResults ports的结果
func PortsResults(responseList []*retryable.Response) []structs.Ports {
	results := make([]structs.Ports, 0)
	for _, response := range responseList {
		var port = structs.Ports{}
		remoteAddr := response.NetConn.Conn.RemoteAddr().String()
		resList := response.ResponseList
		for _, res := range resList {
			port.Name = Name()
			port.Value = Value(remoteAddr)
			port.Method = Method()
			port.State = State(resList)
			port.Product = Product(res)
			port.Version = Version(res)
			port.Iwp = Iwp(remoteAddr)
		}
		results = append(results, port)

	}
	return results
}

func Name() string {
	return "web"
}

func Value(remoteAddr string) string {
	value := strings.Split(remoteAddr, ":")
	return value[1]

}

func Method() string {
	return "brute"
}

func State(responseList []*http.Response) string {
	if len(responseList) != 0 {
		return "open"
	} else {
		return "close"
	}
}

func Product(response *http.Response) string {
	server := response.Header.Get("Server")
	reg := regexp.MustCompile(`[a-zA-Z-_\.]+`)
	return reg.FindString(server)
}

func Version(response *http.Response) string {
	server := response.Header.Get("Server")
	reg := regexp.MustCompile(`[0-9-_./\s]+`)
	return reg.FindString(server)
}

func Iwp(remoteAddr string) string {
	return remoteAddr
}
