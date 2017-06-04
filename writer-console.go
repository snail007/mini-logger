package logger

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type ConsoleWriter struct {
	Format string
}

func NewDefaultConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{
		Format: "[{level}] [{date} {time}.{mili}] {text}",
	}
}
func (w *ConsoleWriter) Write(e Entry) {
	date := time.Unix(e.Timestamp, 0).Format("2006/01/02")
	time := time.Unix(e.Timestamp, 0).Format("15:04:05")
	c := strings.Replace(w.Format, "{level}", e.getLevelString(), -1)
	c = strings.Replace(c, "{date}", date, -1)
	c = strings.Replace(c, "{time}", time, -1)
	c = strings.Replace(c, "{mili}", strconv.FormatInt(e.Milliseconds, 10), -1)
	c = strings.Replace(c, "{text}", e.Content, -1)
	fmt.Fprintln(os.Stdout, c)
}
