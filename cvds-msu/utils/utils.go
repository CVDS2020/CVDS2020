package utils

import (
	"errors"
	"github.com/eiannone/keyboard"
	"log"
	"os"
)

func PauseExit() {
	log.Println("Press any to exit")
	keyboard.GetSingleKey()
	os.Exit(0)
}

var ErrNotDirector = errors.New("file not directory")

func EnsureDir(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(path, 0644); err != nil {
				return err
			}
		}
	} else if !stat.IsDir() {
		return ErrNotDirector
	}
	return nil
}
