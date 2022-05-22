package utils

import (
	"io"
	"log"
	"os"
)

var StdLog *log.Logger
var ErrLog *log.Logger
var InfoLog *log.Logger

func LoggingSetting(fileName string) {
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	logWriter := io.MultiWriter(os.Stdout, f)
	StdLog = log.New(f, "[Std] ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrLog = log.New(f, "[Error] ", log.Ldate|log.Ltime|log.Lshortfile)
	InfoLog = log.New(f, "[Info] ", log.Ldate|log.Ltime|log.Lshortfile)
	StdLog.SetOutput(logWriter)
	ErrLog.SetOutput(logWriter)
	InfoLog.SetOutput(logWriter)
}
