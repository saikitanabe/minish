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

	"github.com/pkg/errors"
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

	minified, err := minifyFiles(src, to, *cssMode)
	if err != nil {
		log.Fatalln("minify failed", err)
	}

	doHash := func() (string, error) {

		if minified.rename {
			return hashRename(minified.target)
		}

		return minified.target, nil
	}

	hashedName, err := doHash()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Minified:", filepath.Base(hashedName))
}

func resolveDir(to string) (string, error) {
	dir := to

	stat, err := os.Stat(to)
	switch {
	case err == nil && stat.IsDir():
	default:
		// this is most probably a file, resolve dir from file path
		dir = filepath.Dir(to)
	}

	stat, _ = os.Stat(dir)
	if stat.IsDir() {
		// still check if dir
		return dir, nil
	}

	return "", fmt.Errorf("Failed to resolve dir from %s", to)
}

func removeFiles(to, filenameEnding string) error {
	log.Println("removeFiles", to, filenameEnding)

	files, err := ioutil.ReadDir(to)
	if err != nil {
		log.Println("ERROR: Failed to read dir", to)
		return errors.Wrapf(err, "removeFiles: failed to read dir %s", to)
	}

	regExp := fmt.Sprintf(`(^\S+-%s)`, strings.Replace(filenameEnding, ".", "\\.", -1))
	fileRegExp := regexp.MustCompile(regExp)

	for _, file := range files {
		if fileRegExp.MatchString(file.Name()) {
			fullpath := fmt.Sprintf("%s/%s", to, file.Name())
			err := os.Remove(fullpath)
			if err != nil {
				return errors.Wrapf(err, "removeFiles: remove failed %s", fullpath)
			}

			fmt.Println("Removed:", file.Name())
		}
	}
	return err
}

func minifyFiles(src, to string, css bool) (*minifiedResult, error) {
	if isMultipleFiles(src) {
		minifiedFiles, err := minifyMultipleFiles(src, to, css)
		if err != nil {
			return nil, errors.Wrapf(err, "minifyFiles: minify multiple files src %s to %s", src, to)
		}

		return concatFiles2(minifiedFiles, to)
	}

	minified, err := minify(src, to, css)
	if err != nil {
		return nil, errors.Wrapf(err, "minifyFiles: src %s to %s", src, to)
	}

	return minified, nil
}

func minifyMultipleFiles(src, to string, css bool) ([]string, error) {
	var result []string

	err := removeMinifiedVersion(to, to)
	if err != nil {
		return nil, err
	}

	for _, file := range strings.Split(src, ",") {
		dir, err := resolveDir(to)
		if err != nil {
			return nil, err
		}
		minified, err := minify(file, dir, css)
		if err != nil {
			return nil, errors.Wrapf(err, "minifyMultipleFiles: minify")
		}
		result = append(result, minified.target)
	}
	return result, nil
}

func removeMinifiedVersion(src, to string) error {
	log.Println("removeMinifiedVersion", src, to)

	_, minfilename, err := targetName(src, to, ".min")
	if err != nil {
		return err
	}

	dir, err := resolveDir(to)
	if err != nil {
		return err
	}

	err = removeFiles(dir, minfilename)
	if err != nil {
		return errors.Wrapf(err, "removeMinifiedVersion: remove files %s %s", dir, minfilename)
	}

	return nil
}

type minifiedResult struct {
	target string
	rename bool
}

func minify(src, to string, css bool) (*minifiedResult, error) {
	fmt.Printf("Minifying %s to %s\n", src, to)

	getTargetName := func() (*minifiedResult, error) {
		if css && strings.HasSuffix(to, ".css") {
			return &minifiedResult{target: to, rename: false}, nil
		}

		target, minfilename, err := targetName(src, to, ".min")
		log.Println("minify:", target, minfilename)

		if err != nil {
			return nil, errors.Wrapf(err, "minify: target name src %s to %s", src, to)
		}

		err = removeFiles(to, minfilename)
		if err != nil {
			return nil, errors.Wrapf(err, "minify: remove files to %s", to)
		}

		return &minifiedResult{target: target, rename: true}, nil
	}

	result, err := getTargetName()
	if err != nil {
		return nil, errors.Wrapf(err, "get target name")
	}

	cmd := func() *exec.Cmd {
		switch {
		case css:
			// --skip-rebase do not rewrite css relative urls
			// this could be a desired option too...
			// return exec.Command("cleancss", "--skip-rebase", src, "-o", result.target)
			return exec.Command("cleancss", src, "-o", result.target)
		default:
			return exec.Command("uglifyjs", src, "-o", result.target, "-c", "-m")
		}
	}()

	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(output))
		return nil, errors.Wrapf(err, "minify: cmd failed %v", cmd)
	}

	if _, err := os.Stat(result.target); os.IsNotExist(err) {
		fmt.Println("Failed to create minified file", result.target)
		return nil, errors.Wrapf(err, "minify: failed to create minified file %s", result.target)
	}

	return result, err
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

func concatFiles2(files []string, to string) (*minifiedResult, error) {
	f, err := os.OpenFile(to, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}

		if _, err = f.WriteString(string(data)); err != nil {
			return nil, err
		}
	}
	return &minifiedResult{target: to, rename: true}, nil
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
