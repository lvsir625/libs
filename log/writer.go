package log

import (
	"io"
	lfile "libs/log/file"
)

type LogWriterType int
const (
	Log_Writer_Type_File LogWriterType = 1 + iota
	Log_Writer_Type_Redis
	Log_Writer_Type_Net_Tcp
)

type LogWriter interface {
	io.WriteCloser
}

func CreateFileLog(fileName string, maxSize int64) LogWriter {
	return &lfile.Writer{
		MaxSize: maxSize,
		FileName: fileName,
	}
}