package main

import (
	"os"

	"github.com/ray-g/dnsproxy/dnsproxy"
)

func main() {
	dnsproxy.Serve(os.Args[1])
}
