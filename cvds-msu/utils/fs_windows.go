package utils

import (
	"os"
	"syscall"
	"time"
)

func GetFileCreateTime(fileInfo os.FileInfo) time.Time {
	sys := fileInfo.Sys().(*syscall.Win32FileAttributeData)
	nano := sys.CreationTime.Nanoseconds()
	return time.Unix(0, nano)
}
