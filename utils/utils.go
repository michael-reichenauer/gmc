package utils

import (
	"log"
	"os"
	"regexp"
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

func CompileRegexp(regexpText string) *regexp.Regexp {
	exp, err := regexp.Compile(regexpText)
	if err != nil {
		log.Fatalf("Failed to compile regexp %q, %v", regexpText, err)
	}
	return exp
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
