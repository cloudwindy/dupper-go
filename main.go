package main

import (
	"fmt"
	"flag"
	"strings"
)

// Non-Case-Sensitive String
type ncstr string

func (a ncstr)eq(b string) bool {
	return strings.EqualFold(string(a), b)
}

var global map[string]interface{}

func init() {
	global = make(map[string]interface{})
}
func main() {
	var ipt string
	fmt.Println("")
	fmt.Scanln(&ipt)
	strings.Split(ipt)
	if ipt.eq("GetEnv"){
		fmt.Println("You are doing right")
	}
	return
}

