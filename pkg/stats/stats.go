package stats

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/projectdiscovery/clistats"
	"httpBruter/pkg/options"
	"log"
	"time"
)

func Statics(args *options.Args) {
	statics, err1 := clistats.New()
	if err1 != nil {
		log.Fatal(err1)
	}

	statics.AddCounter("Request", 0)
	statics.AddCounter("Error", 0)
	statics.AddCounter("Screenshot", 0)
	statics.AddCounter("Fallback", 0)
	statics.AddCounter("Match", 0)
	statics.AddCounter("Database", 0)
	statics.AddStatic("startedAt", time.Now())
	args.Clistats = statics

	//printMutex := &sync.Mutex{}
	var second, minute, hour int
	err := statics.Start(func(stats clistats.StatisticsClient) {
		requestCount, _ := stats.GetCounter("Request")
		errorCount, _ := stats.GetCounter("Error")
		screenshotCount, _ := stats.GetCounter("Screenshot")
		fallbackCount, _ := stats.GetCounter("Fallback")
		startedAt, _ := stats.GetStatic("startedAt")
		matchCount, _ := statics.GetCounter("Match")
		databaseCount, _ := statics.GetCounter("Database")

		runtime := (time.Now().UnixNano() - startedAt.(time.Time).UnixNano()) / 1e9
		second = int(runtime) % 60
		if runtime >= 60 {
			minute = int(runtime) / 60
		}
		if minute >= 60 {
			hour = minute % 60
			minute = 0
		}

		percent := float64(requestCount) / float64(args.RequestsCount) * 100
		data := fmt.Sprintf("[INFO] [%d:%d:%d] Requests: [%d/%d](%.2f%s)",
			hour, minute, second, requestCount, args.RequestsCount, percent, "%")
		if args.Screenshot {
			data = fmt.Sprintf(" %s Screenshot: [%d]", data, screenshotCount)
		}
		if args.NoFallback {
			data = fmt.Sprintf(" %s Fallback: [%d]", data, fallbackCount)
		}
		if args.Match != "" {
			data = fmt.Sprintf(" %s MatchCount: [%d]", data, matchCount)
		}
		if args.Database {
			data = fmt.Sprintf(" %s DatabaseCount: [%d]", data, databaseCount)
		}
		data = fmt.Sprintf(" %s Errors: [%d]", data, errorCount)

		//printMutex.Lock()
		color.Yellow("%s\r\n", data)
		//printMutex.Unlock()
	}, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}

}
