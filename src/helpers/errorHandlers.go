package helpers

import (
	"fmt"
	"log"
	"runtime/debug"
)

func HandleError(err error) {
	st := debug.Stack()
	strace := fmt.Sprintf("%+v \n %+v", err.Error(), string(st))
	log.Output(2, strace)
}
