package helpers

import (
	"fmt"
	"log"
	"runtime/debug"
)

func HandleError(err error) {
	st := debug.Stack()
	strace := fmt.Sprintf("%v \n %v", err, st)
	log.Output(1, strace)
}
