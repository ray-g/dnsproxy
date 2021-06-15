package main

import (
	"os"

	"github.com/ray-g/dnsproxy/dnsproxy"
	"github.com/ray-g/dnsproxy/utils"
)

func main() {
	dnsproxy.Serve(os.Args[1])
	utils.WaitSysSignal()
}
