package main

import (
	"os"

	"github.com/ray-g/dnsproxy/dnsproxy"
)

func main() {
	dnsproxy.ServeWithConfig(os.Args[1])
}
