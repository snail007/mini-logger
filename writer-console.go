package logger

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

type ConsoleWriter struct {
	Format string
}

func NewDefaultConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{
		Format: "[{level}] [{date} {time}.{mili}] {fields} {text}",
	}
}
func (w *ConsoleWriter) Init() {

}
func (w *ConsoleWriter) Write(e Entry) {
	fg := color.New(color.FgHiWhite).SprintFunc()
	switch e.Level {
	case InfoLevel:
		fg = color.New(color.FgHiGreen).SprintFunc()
	case WarnLevel:
		fg = color.New(color.FgHiYellow).SprintFunc()
	case ErrorLevel:
		fg = color.New(color.FgHiRed).SprintFunc()
	case FatalLevel:
		fg = color.New(color.FgHiRed, color.FgHiMagenta).SprintFunc()
	}
	date := time.Unix(e.Timestamp, 0).Format("2006/01/02")
	time := time.Unix(e.Timestamp, 0).Format("15:04:05")
	c := strings.Replace(w.Format, "{level}", e.LevelString, -1)
	c = strings.Replace(c, "{date}", date, -1)
	c = strings.Replace(c, "{time}", time, -1)
	c = strings.Replace(c, "{fields}", e.FieldsString, -1)
	c = strings.Replace(c, "{mili}", strconv.FormatInt(e.Milliseconds, 10), -1)
	c = strings.Replace(c, "{text}", e.Content, -1)
	fmt.Fprintln(os.Stdout, fg(c))
}
