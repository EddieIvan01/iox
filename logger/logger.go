package logger

import (
	"fmt"
	"iox/option"
	"os"
)

const (
	WARN    = "[!]"
	INFO    = "[+]"
	SUCCESS = "[*]"
)

func Info(format string, args ...interface{}) {
	if option.VERBOSE {
		fmt.Fprintf(os.Stdout, INFO+" "+format+"\n", args...)
	}
}

func Warn(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, WARN+" "+format+"\n", args...)
}

func Success(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, SUCCESS+" "+format+"\n", args...)
}
