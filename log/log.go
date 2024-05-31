package log

import (
	"fmt"
	"os"
)

func Fatal(v ...any) {
	fmt.Println(v...)
	os.Exit(1)
}
