package main

import (
	// "bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func removeFiles(dir, filenameEnding string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Println("ERROR: Failed to read dir", dir)
		return
	}

	regExp := fmt.Sprintf(`(^\S+-%s)`, strings.Replace(filenameEnding, ".", "\\.", -1))
	fileRegExp := regexp.MustCompile(regExp)

	for _, file := range files {
		if fileRegExp.MatchString(file.Name()) {
			fullpath := fmt.Sprintf("%s/%s", dir, file.Name())
			err := os.Remove(fullpath)
			if err != nil {
				log.Fatal(err)
			} else {
				fmt.Println("Removed:", file.Name())
			}
		}
	}
}

func minify(src, to string) (string, error) {
	filename := filepath.Base(src)
	ext := filepath.Ext(src)
	name := strings.TrimSuffix(filename, ext)
	minfilename := name + ".min" + ext
	target := fmt.Sprintf("%s/%s", to, minfilename)

	removeFiles(to, minfilename)
	cmd := exec.Command("uglifyjs", src, "-o", target, "-c", "-m")

	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(output))
	}

	return target, err
}

func hash(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	hasher := md5.New()
	hasher.Write(data)

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func hashRename(path string) (string, error) {
	h, err := hash(path)
	if err != nil {
		return "", err
	}

	filename := filepath.Base(path)
	fpath := filepath.Dir(path)
	newname := fmt.Sprintf("%s/%s-%s", fpath, h, filename)
	return newname, os.Rename(path, newname)
}

func usage() {
	fmt.Println(`minify <unminified javascript file name> <output dir>`)
}

func main() {
	if len(os.Args) != 3 {
		usage()
		return
	}
	src := os.Args[1]
	to := os.Args[2]

	minified, err := minify(src, to)
	if err != nil {
		log.Fatal(err)
	}

	hashedName, err := hashRename(minified)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Minified:", filepath.Base(hashedName))
}
