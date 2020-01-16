package utils

import "fmt"

func SetConsoleTitle(title string) (int, error) {
	fmt.Printf("\033]0;%s\007", title)
	return 0, nil
}
