package main

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/panjf2000/ants/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"httpBruter/pkg/database"
	"httpBruter/pkg/finger"
	"httpBruter/pkg/options"
	"httpBruter/pkg/parseCIDR"
	"httpBruter/pkg/parseFile"
	"httpBruter/pkg/runner"
	"httpBruter/pkg/stats"
	"log"
	"os"
	"sync"
)

func banner() {
	color.Green("\n" +
		"██╗  ██╗████████╗████████╗██████╗ ██████╗ ██████╗ ██╗   ██╗████████╗███████╗██████╗ \n" +
		"██║  ██║╚══██╔══╝╚══██╔══╝██╔══██╗██╔══██╗██╔══██╗██║   ██║╚══██╔══╝██╔════╝██╔══██╗\n" +
		"███████║   ██║      ██║   ██████╔╝██████╔╝██████╔╝██║   ██║   ██║   █████╗  ██████╔╝\n" +
		"██╔══██║   ██║      ██║   ██╔═══╝ ██╔══██╗██╔══██╗██║   ██║   ██║   ██╔══╝  ██╔══██╗\n" +
		"██║  ██║   ██║      ██║   ██║     ██████╔╝██║  ██║╚██████╔╝   ██║   ███████╗██║  ██║\n" +
		"╚═╝  ╚═╝   ╚═╝      ╚═╝   ╚═╝     ╚═════╝ ╚═╝  ╚═╝ ╚═════╝    ╚═╝   ╚══════╝╚═╝  ╚═╝\n" +
		"\t\t\t\t\t\t\t\tVersion : 2.0\n\n")
}

func main() {
	banner()
	arg := options.InitArg()
	_, _ = fmt.Fprintf(color.Output, "%s %s | %s | %s | %s \n",
		color.YellowString("%s", "[Params]"),
		color.CyanString("%s", "NormalTimeout:[15s]"),
		color.MagentaString("%s", "HeadlessTimeout:[30s]"),
		color.RedString("%s", fmt.Sprintf("Retries:[%d]", arg.Retries)),
		color.BlueString("%s", fmt.Sprintf("Threads:[%d]", arg.Threads)),
	)
	//创建一个全局的CIDR map列表，每次reverseIP时从这里先查找
	arg.CIDRMap = make(map[string][]string)

	color.HiGreen("[ %-30s | %-40s | %-15s | %-10s | %-10s] \n", "URL", "Title", "FingerPrint", "StatuCode", "Content-Length")

	//判断是stdin 还是输入的文件
	readList := parseFile.Stdin()
	if len(readList) == 0 {
		readList = parseFile.ReadFile(arg.IL)
	}

	requestList := parseCIDR.ParseCIDR(readList)
	arg.RequestsCount = len(requestList)
	arg.FingerContent, arg.FingerFileBuffer = finger.OpenFinger()

	//连接数据库
	if arg.Database {
		arg.MongoDB, arg.MongoClient = database.Collect()
		//最后关闭连接
		defer func(MongoClient *mongo.Client, ctx context.Context) {
			err := MongoClient.Disconnect(ctx)
			if err != nil {
				log.Fatal(err)
			}
		}(arg.MongoClient, context.TODO())
	}

	//最后关闭指纹文件
	defer func(FingerFileBuffer *os.File) {
		err := FingerFileBuffer.Close()
		if err != nil {
			fmt.Println("关闭指纹文件失败")
			os.Exit(-1)
		}
	}(arg.FingerFileBuffer)

	if arg.OT != "" {
		arg.WriteResults, _ = os.OpenFile(arg.OT, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0664)
		defer func(WriteResults *os.File) {
			err := WriteResults.Close()
			if err != nil {
				log.Fatal("[Error] 关闭txt文件失败")
			}
		}(arg.WriteResults)
	}
	if arg.OJ != "" {
		arg.WriteJsonResults, _ = os.OpenFile(arg.OJ, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0664)
		defer func(WriteJsonResults *os.File) {
			err := WriteJsonResults.Close()
			if err != nil {
				log.Fatal("[Error] 关闭json文件失败")
			}
		}(arg.WriteJsonResults)
	}
	//输出统计信息
	if arg.Stats {
		stats.Statics(arg)
	}

	NormalPool(requestList, arg)

}
func NormalPool(requestList []string, arg *options.Args) {
	var wg sync.WaitGroup

	p, _ := ants.NewPoolWithFunc(arg.Threads, func(s interface{}) {
		runner.Runner(s.(string), arg)
		wg.Done()
	})
	defer p.Release()
	for i := 0; i < len(requestList); i++ {
		wg.Add(1)
		_ = p.Invoke(requestList[i])
	}
	wg.Wait()
}
