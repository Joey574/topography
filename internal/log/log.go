package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

var ws = []io.WriteCloser{io.WriteCloser(os.Stdout)}

func PushLogFiles(logs []string) {
	w := make([]io.WriteCloser, 0, len(logs))
	for i := range logs {
		if f, _ := os.OpenFile(logs[i], os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600); f != nil {
			w = append(w, f)
		}
	}

	ws = append(ws, w...)
}

func Close() {
	for _, w := range ws {
		w.Close()
	}
}

func Logf(format string, a ...any) {
	if ws == nil {
		return
	}

	str := fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), format)
	for _, w := range ws {
		fmt.Fprintf(w, str, a...)
	}
}
