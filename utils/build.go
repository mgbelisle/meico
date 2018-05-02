package main

import (
	"flag"
	"fmt"
	// "log"
	"os"
	"path/filepath"
)

var usagePrefix = fmt.Sprintf(`Builds a static site using go templates

Usage: %s [OPTIONS]

OPTIONS:
`, os.Args[0])

var (
	inFlag      = flag.String("in", ".", "Input dir")
	outFlag     = flag.String("out", "dist", "Output dir")
	verboseFlag = flag.Bool("verbose", false, "Verbose output")
)

func main() {
	// Flag setup
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usagePrefix)
		flag.PrintDefaults()
	}
	flag.Parse()

	dir := *inFlag
	subDirToSkip := "skip" // dir/to/walk/skip

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", dir, err)
			return err
		}
		if info.IsDir() && info.Name() == subDirToSkip {
			fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
			return filepath.SkipDir
		}
		fmt.Printf("visited file: %q\n", path)
		return nil
	})

	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", dir, err)
	}
}
