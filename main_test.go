package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"testing"
)

func TestMinifyJavaScriptFile(t *testing.T) {
	doTest(
		t,
		[]string{"minish", "example.js", "dist"},
		regexp.MustCompile(`^\S+-example\.min\.js$`),
	)
}

func TestMinifyConcatenatedJavaScriptFiles(t *testing.T) {
	doTest(
		t,
		[]string{"minish", "example.js,second.js", "dist/bundle.min.js"},
		regexp.MustCompile(`^\S+-bundle\.min\.js$`),
	)
}

func TestMinifyCssFile(t *testing.T) {
	doTest(
		t,
		[]string{"minish", "-css", "example.css", "dist"},
		regexp.MustCompile(`^\S+-example\.min\.css$`),
	)
}

func TestMinifyCssAsNamedVersion(t *testing.T) {
	doTest(
		t,
		[]string{"minish", "-css", "example.css", "dist/example-1.0.min.css"},
		regexp.MustCompile(`^example-1\.0\.min\.css$`),
	)
}

func doTest(t *testing.T, args []string, fileRegExp *regexp.Regexp) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	err := os.RemoveAll("dist")
	if err != nil {
		t.Error("Failed to remove dist", err)
		return
	}

	err = os.Mkdir("dist", os.ModePerm)
	if err != nil {
		t.Error("Failed to create dir dist", err)
		return
	}

	os.Args = args
	main()

	found, err := FindLatest("dist", fileRegExp)
	if err != nil {
		t.Error("command failed", os.Args)
	}

	fmt.Println("found", found)
}

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
