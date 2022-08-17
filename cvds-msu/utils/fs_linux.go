package utils

import (
	"os"
	"syscall"
	"time"
)

func GetFileCreateTime(fileInfo os.FileInfo) time.Time {
	sys := fileInfo.Sys().(*syscall.Stat_t)
	nano := sys.Atim.Nano()
	return time.Unix(0, nano)
}
