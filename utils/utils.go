package utils

import (
	"encoding/json"
	"fmt"
	"github.com/michael-reichenauer/gmc/utils/log"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

func CurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(log.Fatal(err))
	}
	return dir
}

func DebugWait() {
	var input string
	fmt.Scanln(&input)
}

func BinPath() string {
	name := os.Args[0]
	var err error
	if name[0] == '.' {
		name, err = filepath.Abs(name)
		if err == nil {
			name = filepath.Clean(name)
		}
	} else {
		name, err = exec.LookPath(filepath.Clean(name))
	}
	if err != nil {
		panic(log.Fatal(err))
	}
	return name
}

func GetVolumes() []string {
	var volumes []string
	if runtime.GOOS == "windows" {
		for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
			volume := string(drive) + ":\\"
			f, err := os.Open(string(drive) + ":\\")
			if err == nil {
				volumes = append(volumes, volume)
				f.Close()
			}
		}
		return volumes
	}
	return []string{"/"}
}

func RecentItems(items []string, item string, maxSize int) []string {
	if i := StringsIndex(items, item); i != -1 {
		items = append(items[:i], items[i+1:]...)
	}
	if len(items) > maxSize {
		items = items[0:1]
	}
	return append([]string{item}, items...)
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
		panic(log.Fatalf(err, "Failed to compile regexp %q", regexpText))
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
	if os.IsNotExist(err) || info == nil {
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
	if os.IsNotExist(err) || info == nil {
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
