package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	cssMode = flag.Bool("css", false, "minify and hash css file")
)

func usage() {
	fmt.Println(`minify [-css] <unminified file names (comma separated list)> <output>

output:
	- if only 1 file uses that as minified output file name
	- in case of multiple files output needs to end with .js or .css	
`)
}

func main() {
	flag.Parse()

	if *cssMode && len(os.Args) != 4 {
		usage()
		return
	}

	if len(os.Args) != 3 && !*cssMode {
		usage()
		return
	}

	baseIndex := func() int {
		if *cssMode {
			return 2
		}
		return 1
	}

	src := os.Args[baseIndex()]
	to := os.Args[baseIndex()+1]

	from, err := makeInput(src, to)
	if err != nil {
		log.Fatalln("Failed to make input", err)
	}

	minified, err := minify(from, to, *cssMode)
	if err != nil {
		log.Fatal(err)
	}

	hashedName, err := hashRename(minified)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Minified:", filepath.Base(hashedName))
}

func resolveDir(to string) (string, error) {
	dir := to
	stat, err := os.Stat(to)
	if err != nil {
		return "", err
	}

	if !stat.IsDir() {
		// this is a file, so resolve base directory
		filename := filepath.Base(to)
		dir = strings.TrimSuffix(to, string(filepath.Separator)+filename)
	}
	return dir, nil
}

func removeFiles(to, filenameEnding string) error {
	dir, err := resolveDir(to)
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Println("ERROR: Failed to read dir", dir)
		return err
	}

	regExp := fmt.Sprintf(`(^\S+-%s)`, strings.Replace(filenameEnding, ".", "\\.", -1))
	fileRegExp := regexp.MustCompile(regExp)

	for _, file := range files {
		if fileRegExp.MatchString(file.Name()) {
			fullpath := fmt.Sprintf("%s/%s", dir, file.Name())
			err := os.Remove(fullpath)
			if err != nil {
				return err
			}

			fmt.Println("Removed:", file.Name())
		}
	}
	return err
}

func minify(src, to string, css bool) (string, error) {
	target, minfilename, err := targetName(src, to, ".min")
	if err != nil {
		return "", err
	}

	err = removeFiles(to, minfilename)
	if err != nil {
		return "", err
	}

	cmd := func() *exec.Cmd {
		switch {
		case css:
			return exec.Command("cleancss", src, "-o", target)
		default:
			return exec.Command("uglifyjs", src, "-o", target, "-c", "-m")
		}
	}

	output, err := cmd().CombinedOutput()

	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(output))
	}

	return target, err
}

func targetName(src, to, extra string) (string, string, error) {
	filename := filepath.Base(src)
	ext := filepath.Ext(src)
	name := strings.TrimSuffix(filename, ext)
	minfilename := name + extra + ext

	dir, err := resolveDir(to)
	if err != nil {
		return "", "", err
	}

	return fmt.Sprintf("%s/%s", dir, minfilename), minfilename, nil
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

func makeInput(src, to string) (string, error) {
	// by default only one input file
	result := src

	if isMultipleFiles(src) {
		var err error
		err = concatFiles(src, to)
		if err != nil {
			return "", err
		}

		result = to
	}

	return result, nil
}

func isMultipleFiles(src string) bool {
	splitted := strings.Split(src, ",")
	return len(splitted) > 1
}

func concatFiles(src, output string) error {
	ext := filepath.Ext(src)
	concat, err := concatFilesInMemory(src, ext)
	if err != nil {
		return err
	}

	data := []byte(concat)

	if _, err := os.Stat(output); os.IsExist(err) {
		return fmt.Errorf("file already exists, cannot overwrite %s", output)
	}

	err = ioutil.WriteFile(output, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func concatFilesInMemory(src, outputExt string) (string, error) {
	result := ""
	fmt.Println("Concateniting")

	for _, file := range strings.Split(src, ",") {
		fmt.Println(file)

		data, err := ioutil.ReadFile(file)
		if err != nil {
			return "", err
		}

		ext := filepath.Ext(file)
		if ext != outputExt {
			return "", fmt.Errorf(
				"Extension differs for file %s output ext %s",
				file,
				outputExt,
			)
		}

		result += string(data)
	}
	return result, nil
}
