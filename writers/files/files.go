package files

import (
	"encoding/json"
	"strings"
	"time"

	"bufio"

	"os"

	"path"

	"compress/gzip"
	"fmt"

	"io/ioutil"

	"github.com/snail007/mini-logger"
)

const (
	T_JSON = iota
	T_TEXT = iota
)

type FileConfig struct {
	IsRotate    bool
	MaxBytes    int64
	MaxCount    int
	LogPath     string
	FileNameSet map[string]uint8
	IsCompress  bool
	Type        int
	Format      string
}
type bufWriter struct {
	file   *os.File
	writer *bufio.Writer
}
type FileWriter struct {
	filePtrMap map[string]*bufWriter
	c          FileConfig
}

func New(config FileConfig) *FileWriter {
	return &FileWriter{
		c:          config,
		filePtrMap: map[string]*bufWriter{},
	}
}

func NewDefault() *FileWriter {
	return &FileWriter{
		filePtrMap: map[string]*bufWriter{},
		c:          GetDefaultFileConfig(),
	}
}

//GetDefaultFileConfig  config for new FileWriter
//IsRotate : whether to rotate log file.
//MaxBytes : when RotateType is file.T_SIZE ,this need to be set.
//MaxCount : how many log files can be remain.
//LogPath  : the folder which store log files,must be exists.
//FileNameSet : key is filename of log file,no extesion.value is levels
//IsCompress : whether to compress rotate log file.
//Type:output format,should be files.T_JSON or files.T_TEXT
//Format:when Type is files.T_TEXT,this  can setting output text format,
//       default is : [{level}] [{date} {time}.{mili}] {text} {fields}
func GetDefaultFileConfig() FileConfig {
	return FileConfig{
		IsRotate: true,
		MaxBytes: 100 * 1024 * 1024,
		MaxCount: 10,
		LogPath:  "log",
		FileNameSet: map[string]uint8{
			"info":  logger.InfoLevel,
			"error": logger.WarnLevel | logger.ErrorLevel | logger.FatalLevel,
		},
		IsCompress: true,
		Type:       T_JSON,
		Format:     "[{level}] [{date} {time}.{mili}] {text} {fields}",
	}
}
func getFilePath(logPath, filename string) string {
	return path.Join(logPath, filename+".log")

}
func (w *FileWriter) Init() (err error) {
	if _, err := os.Stat(w.c.LogPath); os.IsNotExist(err) {
		os.Mkdir(w.c.LogPath, 0700)
	}
	for filename := range w.c.FileNameSet {
		f, err := os.OpenFile(getFilePath(w.c.LogPath, filename), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			return err
		}
		bf := bufio.NewWriter(f)
		w.filePtrMap[filename] = &bufWriter{
			file:   f,
			writer: bf,
		}
	}
	return
}
func (w *FileWriter) Write(e logger.Entry) {
	for filename, levels := range w.c.FileNameSet {
		if e.Level&levels == e.Level {
			date := time.Unix(e.Timestamp, 0).Format("2006/01/02")
			time := time.Unix(e.Timestamp, 0).Format("15:04:05")
			var c string
			if w.c.Type == T_TEXT {
				c = strings.Replace(w.c.Format, "{level}", fmt.Sprintf("%-5s", e.LevelString), -1)
				c = strings.Replace(c, "{date}", date, -1)
				c = strings.Replace(c, "{time}", time, -1)
				c = strings.Replace(c, "{fields}", e.FieldsString, -1)
				c = strings.Replace(c, "{mili}", fmt.Sprintf("%03d", e.Milliseconds), -1)
				c = strings.Replace(c, "{text}", e.Content, -1)
			} else if w.c.Type == T_JSON {
				e.LevelString = strings.TrimRight(e.LevelString, " ")
				e.FieldsString = ""
				v, _ := json.Marshal(e)
				c = string(v)
			} else {
				return
			}
			w.filePtrMap[filename].writer.WriteString(fmt.Sprintln(c))
			w.filePtrMap[filename].writer.Flush()
			if w.c.IsRotate {
				var filepath = getFilePath(w.c.LogPath, filename)
				if stat, err := os.Stat(filepath); err == nil {
					if stat.Size() > w.c.MaxBytes {
						w.filePtrMap[filename].file.Close()
						var i = 0
						for {
							if ok, err := exists(getFilePath(w.c.LogPath, fmt.Sprintf("%s.%d", filename, i))); (err == nil && !ok) || err != nil {
								if err != nil {
									fmt.Printf("ERROR fail to get stat of log file,%s", err)
									return
								}
								break
							}
							i++
						}
						newfilepath := getFilePath(w.c.LogPath, fmt.Sprintf("%s.%d", filename, i))
						e := os.Rename(filepath, newfilepath)
						if e != nil {
							fmt.Printf("ERROR fail to Rename log file,%s", e)
							return
						}
						go func() {
							f, _ := os.OpenFile(newfilepath+".gz", os.O_CREATE|os.O_WRONLY, 0600)
							rf, _ := os.OpenFile(newfilepath, os.O_RDONLY, 0600)
							gz, err := gzip.NewWriterLevel(f, gzip.BestCompression)
							b, _ := ioutil.ReadAll(rf)
							if err == nil {
								gz.Write(b)
								gz.Flush()
								gz.Close()
								f.Close()
								rf.Close()
								os.Remove(newfilepath)
							} else {
								fmt.Printf("ERROR fail to compress log file,%s", err)
							}
						}()
						f, e := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
						if e != nil {
							fmt.Println("ERROR fail to create log file")
							return
						}
						bf := bufio.NewWriter(f)
						w.filePtrMap[filename] = &bufWriter{
							file:   f,
							writer: bf,
						}
					}
				}
			}
		}
	}
}
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
