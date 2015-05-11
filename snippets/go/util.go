package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
)

func FindLatest(dir string, fileRegExp *regexp.Regexp) (string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var latest int64 = 0
	thefile := ""

	for _, file := range files {
		if fileRegExp.MatchString(file.Name()) && file.ModTime().UnixNano() > latest {
			latest = file.ModTime().UnixNano()
			thefile = file.Name()
		}
	}

	if latest == 0 {
		return "", fmt.Errorf("File not found %s/%+v", dir, fileRegExp)
	}

	return thefile, nil
}

func main() {
	latest, err := FindLatest("../../dist", regexp.MustCompile(`^\S+-example\.min\.js$`))

	if err != nil {
		fmt.Println("Failed to find latest", err)
		return
	}
	fmt.Println("Latest", latest)
}
