package main

import (
	"fmt"
	"github.com/CVDS2020/CVDS2020/common/assert"
	"github.com/CVDS2020/CVDS2020/cvds-msu/utils"
	"os"
)

func main() {
	fmt.Println(utils.GetFileCreateTime(assert.Must(os.Stat(os.Args[1]))))
}
