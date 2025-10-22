package verbose

import (
	"fmt"
	"log"
	"os"
)

var Verbose bool

func VerbosePrintln(a ...interface{}) {
	if Verbose {
		fmt.Println(a...)
	}
}

func VerbosePrintf(format string, a ...interface{}) {
	if Verbose {
		log.Printf(format, a...)
	}
}

func Printf(format string, a ...interface{}) {
	log.Printf(format, a...)
}

func VerboseErrorf(format string, a ...interface{}) error {
	msg := fmt.Sprintf(format, a...)
	log.Println("[ERROR]", msg)
	return fmt.Errorf(msg)
}

func VerboseFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func VerboseFatalfMsg(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	log.Println("[FATAL]", msg)
	os.Exit(1)
}
