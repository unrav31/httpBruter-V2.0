package stdin

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Stdin stdin模式，接收参数和输入
func Stdin() []string {
	readList := make([]string, 0)
	stat, err := os.Stdin.Stat()
	if err != nil {
		return readList
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
	return readList
}

type Judge struct {
	OrEqual     string
	OrNotEqual  string
	AndEqual    string
	AndNotEqual string
	Equal       string
	NotEqual    string
	Flag        string
	Result      bool
}

var judge Judge

// MatchStr 提取输入的正则表达式和表达式类型(与、或)
func MatchStr(stdin string) Judge {
	if strings.Contains(stdin, "||") {
		orList := strings.Split(stdin, "||") //包含或运算，提取||两边的字符串
		for i := 0; i < len(orList); i++ {
			if strings.Contains(orList[i], "title!=") { //包含title不等于,提取这里面的正则表达式，在解析结果中只要匹配，就不加入截图列表
				Reg := regexp.MustCompile(`title!='(.*?)'`)
				regxStr := Reg.FindStringSubmatch(orList[i])
				if len(regxStr) > 1 && regxStr[1] != "" {
					judge.OrNotEqual = regxStr[1]
				} else {
					fmt.Println("judge.OrNotEqual==nil")
				}

			} else if strings.Contains(orList[i], "title=") { //包含title等于,提取这里面的正则表达式，在解析结果中只要匹配，就加入截图列表
				Reg := regexp.MustCompile(`title='(.*?)'`)
				regStr := Reg.FindStringSubmatch(orList[i])
				if len(regStr) > 1 && regStr[1] != "" {
					judge.OrEqual = regStr[1]
				} else {
					fmt.Println("judge.OrEqual==nil")
				}
			} else { //既不不等于，也不等于，就继续读取，这种情况一般不会出现
				continue
			}
		}
		judge.Flag = "OR"

	} else if strings.Contains(stdin, "&&") {
		andList := strings.Split(stdin, "&&") //包含与运算
		for i := 0; i < len(andList); i++ {
			if strings.Contains(andList[i], "title!=") { //包含title不等于
				Reg := regexp.MustCompile(`title!='(.*?)'`)
				regxStr := Reg.FindStringSubmatch(andList[i])
				if len(regxStr) > 1 && regxStr[1] != "" {
					judge.AndNotEqual = regxStr[1]
				} else {
					fmt.Println("judge.AndNotEqual==nil")
				}
			} else if strings.Contains(andList[i], "title=") { //包含title等于
				Reg := regexp.MustCompile(`title='(.*?)'`)
				regxStr := Reg.FindStringSubmatch(andList[i])
				if len(regxStr) != 0 && regxStr[1] != "" {
					judge.AndEqual = regxStr[1]
				} else {
					fmt.Println("judge.AndEqual==nil")
				}
			} else { //既不不等于，也不等于，就继续读取，这种情况一般不会出现
				continue
			}
		}
		judge.Flag = "AND"
	} else {                                    //没有||、&&,直接解析
		if strings.Contains(stdin, "title!=") { //包含title不等于
			Reg := regexp.MustCompile(`title!='(.*?)'`)
			regxStr := Reg.FindStringSubmatch(stdin)
			if len(regxStr) != 0 && regxStr[1] != "" {
				judge.NotEqual = regxStr[1]
			}
		} else if strings.Contains(stdin, "title=") { //包含title等于
			Reg := regexp.MustCompile(`title='(.*?)'`)
			regxStr := Reg.FindStringSubmatch(stdin)
			if len(regxStr) != 0 && regxStr[1] != "" {
				judge.Equal = regxStr[1]
			}
		} else { //什么都没匹配到
			return judge
		}
		judge.Flag = "NIL"
	}

	return judge
}

// Match
func Match(matchRegx string, title string) bool {
	judge = MatchStr(matchRegx)
	var first = false
	var second = false
	var final = false
	if judge.Flag == "OR" { //||
		if judge.OrNotEqual != "" { //title!=
			first, _ = regexp.MatchString(judge.OrNotEqual, title) //true表示有title!=bad title,需要使false才加入列表
		}
		if judge.OrEqual != "" { //title==
			second, _ = regexp.MatchString(judge.OrEqual, title)
		}
		if !first || second {
			final = true
			return final
		}
	}
	if judge.Flag == "AND" { //&&
		if judge.AndEqual != "" { //title==
			first, _ = regexp.MatchString(judge.AndEqual, title) //true表示有title==good title
		}
		if judge.AndNotEqual != "" { //title!=
			second, _ = regexp.MatchString(judge.AndNotEqual, title) //true表示有title!=bad title,需要使false才加入列表
		}
		if first && !second {
			final = true
			return final
		}
	}

	if judge.Flag == "NIL" { //没有与、或
		if judge.Equal != "" { //title=
			first, _ = regexp.MatchString(judge.Equal, title) //true表示有title==good title
			if first {
				return true
			}
		} else if judge.NotEqual != "" { //title!=
			second, _ = regexp.MatchString(judge.NotEqual, title) //true表示有title!=bad title，需要使false才加入列表
			if !second {
				return true
			}
		}

	}
	return final
}
