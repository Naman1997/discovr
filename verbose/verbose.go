package verbose

import "fmt"

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
