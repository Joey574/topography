package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

var ws = []io.Writer{io.Writer(os.Stdout)}

func SetLogFile(nw io.Writer) {
	ws = append(ws, nw)
}

func Logf(format string, a ...any) {
	if ws == nil {
		return
	}

	stamp := fmt.Sprintf("[%s] ", time.Now().Format("2006-01-02 15:04:05"))
	for _, w := range ws {
		fmt.Fprintf(w, stamp+format+"\n", a...)
	}
}
