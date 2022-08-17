package utils

import (
	"fmt"
	"github.com/CVDS2020/CVDS2020/common/assert"
	"os"
	"testing"
)

func TestFS(t *testing.T) {
	fmt.Println(GetFileCreateTime(assert.Must(os.Stat("fs_windows.go"))))
	a := "%03s"
	fmt.Printf(a, 10)
}
