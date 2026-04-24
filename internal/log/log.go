package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

var w = io.Writer(os.Stdout)

func SetLogFile(nw io.Writer) {
	w = nw
}

func Logf(format string, a ...any) {
	if w == nil {
		return
	}

	stamp := fmt.Sprintf("[%s] ", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, stamp+format+"\n", a...)
}
