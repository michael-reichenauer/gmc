package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
)

func CurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func BinPath() string {
	return os.Args[0]
}

func Text(text string, length int) string {
	textLength := len([]rune(text))
	if textLength < length {
		return text + strings.Repeat(" ", length-textLength)
	}
	return string([]rune(text)[0:length])
}

func RunesText(text string, length int) string {
	textLength := len([]rune(text))
	if textLength < length {
		return text + strings.Repeat(" ", length-textLength)
	}
	return string([]rune(text)[0:length])
}

func CompileRegexp(regexpText string) *regexp.Regexp {
	exp, err := regexp.Compile(regexpText)
	if err != nil {
		log.Fatalf("Failed to compile regexp %q, %v", regexpText, err)
	}
	return exp
}

func MustJsonMarshal(v interface{}) []byte {
	bytes, err := json.Marshal(v)
	if err != nil {
		log.Fatal(err)
	}
	return bytes
}

func MustJsonUnmarshal(bytes []byte, v interface{}) {
	err := json.Unmarshal(bytes, v)
	if err != nil {
		log.Fatal(err)
	}
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
func MustFileWrite(filename string, data []byte) {
	err := FileWrite(filename, data)
	if err != nil {
		log.Fatal(err)
	}
}
func FileWrite(filename string, data []byte) error {
	err := ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
	return err
}
func MustFileRead(filename string) []byte {
	bytes, err := FileRead(filename)
	if err != nil {
		log.Fatal(err)
	}
	return bytes
}
func FileRead(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

func DirExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func StringsContains(s []string, e string) bool {
	return StringsIndex(s, e) != -1
}

func StringsIndex(s []string, e string) int {
	for i, a := range s {
		if a == e {
			return i
		}
	}
	return -1
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
