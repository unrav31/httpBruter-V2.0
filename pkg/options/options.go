package options

import (
	"flag"
	"github.com/projectdiscovery/clistats"
	"go.mongodb.org/mongo-driver/mongo"
	"httpBruter/pkg/finger"
	"os"
)

func InitArg() *Args {
	arg := &Args{}
	flag.StringVar(&arg.IL, "iL", "", "输入文件名")
	flag.IntVar(&arg.Threads, "threads", 50, "设置并发数，如果开启了headless或截图，建议将并发数设置到20以内")
	flag.BoolVar(&arg.Headless, "headless", false, "使用Headless访问")
	flag.BoolVar(&arg.Stats, "stats", false, "显示实时统计信息")
	flag.BoolVar(&arg.NoFallback, "no-fallback", false, "如果IP地址是443端口，再次访问其80端口")
	flag.BoolVar(&arg.Screenshot, "screenshots", false, "开启截图(保存在当前目录的'screenshots'文件夹中)")
	flag.IntVar(&arg.Retries, "retries", 0, "设置重试次数，默认0")
	flag.StringVar(&arg.Binary, "binary", "", "设置chrome浏览器的默认路径")
	flag.BoolVar(&arg.Debug, "debug", false, "开启错误输出")
	flag.StringVar(&arg.OT, "oT", "", "设置输出普通文件的路径")
	flag.StringVar(&arg.Match, "match", "", "使用正则匹配过滤结果")
	flag.StringVar(&arg.OJ, "oJ", "", "设置输出json文件的路径")
	flag.BoolVar(&arg.Cdn, "cdn", false, "开启CDN探测")
	flag.BoolVar(&arg.Vhost, "vhost", false, "开启VHOST探测")
	flag.BoolVar(&arg.Redirects, "no-redirects", false, "设置不跟随重定向")
	flag.BoolVar(&arg.Finger, "finger", false, "开启CMS指纹探测")
	flag.BoolVar(&arg.Database, "database", false, "存入远程mongo数据库")
	flag.BoolVar(&arg.ReverseIP, "reverse", false, "开启反查IP功能")
	flag.StringVar(&arg.ScreenshotPath, "screenshots-path", "", "设置截图存储路径")
	flag.BoolVar(&arg.ScreenshotFullPage, "screenshots-fullpage", false, "设置全屏截图")
	flag.StringVar(&arg.ReverseIPOpt, "reverse-option", "all", "选择反查IP的接口：'ipchaxun'或'webscan'")
	flag.Parse()
	return arg
}

// Args 程序运行参数
type Args struct {
	IL                 string
	Threads            int
	Headless           bool
	Stats              bool
	NoFallback         bool
	Screenshot         bool
	Retries            int
	Binary             string
	HeadlessThreads    int
	Debug              bool
	OT                 string
	Match              string
	OJ                 string
	Cdn                bool
	Vhost              bool
	Redirects          bool
	Finger             bool
	Version            string
	Database           bool
	ReverseIP          bool
	ReverseIPOpt       string
	ScreenshotPath     string
	ScreenshotFullPage bool

	//下面是全局变量
	FingerContent    []finger.Content
	FingerFileBuffer *os.File
	Clistats         *clistats.Statistics
	WriteResults     *os.File
	WriteJsonResults *os.File
	RequestsCount    int
	MongoDB          *mongo.Database
	MongoClient      *mongo.Client
	CIDRMap          map[string][]string
}
