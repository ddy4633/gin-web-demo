package conf

import "fmt"

//检查错误
func CheckERR(err error,info string) bool {
	if err != nil {
		fmt.Printf("%s ERROR->%s\n",info,err)
		return false
	}
	return true
}
