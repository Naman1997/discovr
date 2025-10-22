package verbose

import (
	"fmt"
	"log"
)

var Verbose bool

func VerbosePrintln(a ...interface{}) {
	if Verbose {
		fmt.Println(a...)
	}
}

func VerbosePrintf(format string, a ...interface{}) {
	if Verbose {
		fmt.Printf(format, a...)
	}
}

func Printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

func VerboseFatalf(err error) {
	log.Fatalf(err.Error())
}
