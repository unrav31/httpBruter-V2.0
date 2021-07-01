package parseFile

import (
	"bufio"
	"log"
	"os"
	"strings"
)

// ReadFile 按行读取文件并strip
func ReadFile(fileName string) (result []string) {

	fileBuffer, err := os.Open(fileName)

	if err != nil {
		log.Fatalf("[x] 错误 %s文件打开失败", fileName)
	}

	reader := bufio.NewReader(fileBuffer)
	for {
		line, _, e := reader.ReadLine()
		if e != nil {
			break
		}

		//如果有换行和空白,跳过
		if string(line) == "" || string(line) == " " || string(line) == "\n" {
			continue
		}
		strip := strings.Replace(string(line), " ", "", -1)
		strip = strings.Replace(strip, "\t", "", -1)
		strip = strings.Replace(strip, "\n", "", -1)
		result = append(result, strip)
	}

	return
}

// Stdin stdin模式输入参数
func Stdin() (readList []string) {

	stat, err := os.Stdin.Stat()

	if err != nil {
		log.Fatal("[x] 错误 stdin不接受此类参数")
	}

	mode := stat.Mode()

	isPipedFromChrDev := (mode & os.ModeCharDevice) == 0
	isPipedFromFIFO := (mode & os.ModeNamedPipe) != 0

	s := bufio.NewScanner(os.Stdin)

	if isPipedFromChrDev || isPipedFromFIFO {
		for s.Scan() {
			readList = append(readList, s.Text())
		}
	}
	return
}
