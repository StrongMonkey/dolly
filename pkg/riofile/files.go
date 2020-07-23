package riofile

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rancher/wrangler/pkg/data/convert"

	"gopkg.in/yaml.v3"
)

const (
	defaultRiofile       = "Riofile"
	defaultRiofileAnswer = "Riofile-answers"

	defaultRiofileContent = `
services:
  %s:
    image: ./
    ports: 80:8080/http`

	defaultRiofileContentWithDockerfile = `
services:
  %s:
    build:
      dockerfile: %s
      context: %s
    ports: 80:8080/http`
)

func LoadFileAndAnswer(path string, answerPath string) ([]byte, map[string]string, error) {
	riofile, err := LoadRiofile(path)
	if err != nil {
		return nil, nil, err
	}
	answer, err := LoadAnswer(answerPath)
	if err != nil {
		return nil, nil, err
	}

	return riofile, answer, nil
}

func LoadRiofile(path string) ([]byte, error) {
	if path != "" {
		content, err := readFile(path)
		if err != nil {
			return nil, err
		}
		// named Riofile, has either valid yaml or templating
		var r map[string]interface{}
		if err := yaml.Unmarshal(content, &r); err == nil || bytes.Contains(content, []byte("goTemplate:")) {
			return content, nil
		}
		// named Dockerfile
		return []byte(fmt.Sprintf(defaultRiofileContentWithDockerfile, getCurrentDir(), filepath.Base(path), filepath.Dir(path))), nil
	}
	// assumed Riofile
	if _, err := os.Stat(defaultRiofile); err == nil {
		return ioutil.ReadFile(defaultRiofile)
	}
	// assumed Dockerfile
	return []byte(fmt.Sprintf(defaultRiofileContent, getCurrentDir())), nil
}

func LoadAnswer(path string) (map[string]string, error) {
	if path == "" {
		if _, err := os.Stat(defaultRiofileAnswer); err == nil {
			path = defaultRiofileAnswer
		}
	}
	return readAnswers(path)
}

func getCurrentDir() string {
	workingDir, _ := os.Getwd()
	dir := filepath.Base(workingDir)
	return strings.ToLower(dir)
}

func readFile(file string) ([]byte, error) {
	if file == "-" {
		return ioutil.ReadAll(os.Stdin)
	}
	if strings.HasPrefix(file, "http") {
		resp, err := http.Get(file)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	}
	return ioutil.ReadFile(file)
}

func readAnswers(answersFile string) (map[string]string, error) {
	content, err := readFile(answersFile)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	result := map[string]string{}
	for k, v := range data {
		result[k] = convert.ToString(v)
	}

	return result, nil
}
