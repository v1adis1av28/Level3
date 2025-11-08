package main

import (
	"fmt"
	"os"

	"github.com/v1adis1av28/level3/ImageProcessor/internal/config"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	zlog.InitConsole()
	zlog.SetLevel("debug")
	cfg, err := config.New("../config/dev.yml")
	if err != nil {
		fmt.Println("Error on reading config %v", err)
		os.Exit(1)
	}
	fmt.Println(cfg)
}
