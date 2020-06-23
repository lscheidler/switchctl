package common

import (
	"log"
	"os"
)

type MyLogger struct {
	filename string
	Log      *log.Logger
	file     *os.File
	level    int
}

func NewMyLogger(filename string, level int) *MyLogger {
	mylog := MyLogger{
		filename: filename,
		level:    level,
	}

	if f, err := os.Create(filename); err == nil {
		mylog.file = f
		mylog.Log = log.New(f, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	} else {
		log.Fatal(err)
	}
	return &mylog
}

func (mylog *MyLogger) Close() {
	mylog.file.Close()
}

func (mylog *MyLogger) Println(level int, v ...interface{}) {
	if level >= mylog.level {
		mylog.Log.Println(v)
	}
}
