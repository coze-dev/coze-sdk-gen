package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/coze-dev/coze-sdk-gen/internal/version"
)

func main() {
	showVersion := flag.Bool("version", false, "print version")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.String())
		return
	}

	fmt.Fprintln(os.Stderr, "coze-sdk-gen: implementation is in progress")
	os.Exit(2)
}
