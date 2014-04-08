package main

import (
	"os"
	"fmt"
	"path"
)

func main() {
	pwd,err := os.Getwd()
	if err != nil {
		panic("Could not get the current working directory")
	}
	for _,arg := range os.Args[1:] {
		dir := path.Dir(arg)
		base := path.Base(arg)
		os.Chdir(dir)
		newPwd,err := os.Getwd()
		if err != nil {
			panic("Could not get the current working directory")
		}
		fmt.Printf("%s/%s\n", newPwd, base)
		os.Chdir(pwd)
	}
}
