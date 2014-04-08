package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"
)

func absolute(pwd, path string) string {
	os.Chdir(path)
	newPath, err := os.Getwd()
	if err != nil {
		panic("Could not get the current working directory")
	}
	os.Chdir(pwd)
	return newPath
}

func updir(str string) (string, string) {
	dirs := strings.Split(str, "\n")
	i := len(dirs)
	if i == 1 {
		return "/", "/"
	}
	return strings.Join(dirs[:(i-2)], "/"), dirs[i-1]
}

func toHome(pwd, dir string) string {
	var str string
	usr, err := user.Current()
	if err != nil {
		panic("Could not get the users homedir")
	}
	home := usr.HomeDir
	os.Chdir(dir)
	npwd := os.Getenv("PWD")
	for {
		var s string
		if npwd == "/" {
			os.Chdir(pwd)
			return absolute(pwd, dir)
		}
		if npwd == home {
			if str == "" {
				return dir
			} else {
				return str + "/" + dir
			}
		}
		npwd, s = updir(npwd)
		str = s + "/" + str
	}
}

var abs = flag.Bool("a", false, "Absolute path")

func main() {
	flag.Parse()
	pwd, err := os.Getwd()
	if err != nil {
		panic("Could not get the current working directory")
	}
	for _, arg := range flag.Args()[:] {
		base := path.Base(arg)
		dir := path.Dir(arg)
		if *abs {
			newPath := absolute(pwd, dir)
			fmt.Printf("%s/%s\n", newPath, base)
		} else {
			if strings.HasPrefix(dir, "/") {
				fmt.Printf("%s/%s\n", dir, base)
			} else {
				fmt.Printf("%s/%s\n", toHome(pwd, dir), base)
			}
		}
	}
}
