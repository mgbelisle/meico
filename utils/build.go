package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var usagePrefix = fmt.Sprintf(`Builds a static site using the html/template package, with TemplateData provided.

Usage: %s [OPTIONS]

OPTIONS:
`, os.Args[0])

var (
	inFlag        = flag.String("in", "src", "Input dir")
	outFlag       = flag.String("out", "www", "Output dir")
	templatesFlag = flag.String("templates", "templates", "Templates dir")
	verboseFlag   = flag.Bool("verbose", false, "Verbose output")
	addrFlag      = flag.String("addr", "", "Address to serve output dir, if provided")
)

type TemplateData struct {
	Root string // Relative path to the root url (e.g. "../..")
}

var (
	verboseLogger = log.New(ioutil.Discard, os.Args[0]+": ", log.LstdFlags)
	errLogger     = log.New(os.Stderr, os.Args[0]+": ", log.LstdFlags)
)

func main() {
	// Flag setup
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usagePrefix)
		flag.PrintDefaults()
	}
	flag.Parse()

	// Logger setup
	if *verboseFlag {
		verboseLogger = log.New(os.Stdout, os.Args[0], log.LstdFlags)
	}

	// Build once
	build(errLogger.Panic)
}

func build(logFunc func(...interface{})) {
	// Templates setup
	tmpl := template.Must(template.ParseGlob(filepath.Join(*templatesFlag, "**")))

	// Render the files
	if err := os.RemoveAll(*outFlag); err != nil {
		logFunc(err)
	}
	wg := sync.WaitGroup{}
	if err := filepath.Walk(*inFlag, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logFunc(err)
		}
		rel, err := filepath.Rel(*inFlag, path)
		if err != nil {
			logFunc(err)
		}
		outPath := filepath.Join(*outFlag, rel)
		if info.IsDir() {
			// Make the dir
			verboseLogger.Printf("Creating %s", outPath)
			if err := os.Mkdir(outPath, info.Mode()); err != nil {
				logFunc(err)
			}
		} else {
			// Otherwise parse the file or copy it, whichever is appropriate.
			// Do them all in parallel
			wg.Add(1)
			go func(path string, outPath string, info os.FileInfo) {
				defer wg.Add(-1)
				outFile, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE, info.Mode())
				if err != nil {
					logFunc(err)
				}
				defer outFile.Close()
				root, err := filepath.Rel(path, *inFlag)
				if err != nil {
					logFunc(err)
				}
				if filepath.Ext(path) == ".html" {
					verboseLogger.Printf("Parsing %s", path)
					tmpl2, err := tmpl.Clone()
					if err != nil {
						logFunc(err)
					}
					tmpl2 = template.Must(tmpl2.ParseFiles(path))
					if err := tmpl2.Execute(outFile, &TemplateData{
						Root: root,
					}); err != nil {
						logFunc(err)
					}
				} else {
					verboseLogger.Printf("Copying %s", path)
					inFile, err := os.Open(path)
					if err != nil {
						logFunc(err)
					}
					defer inFile.Close()
					if _, err := io.Copy(outFile, inFile); err != nil {
						logFunc(err)
					}
				}
			}(path, outPath, info)
		}
		return nil
	}); err != nil {
		logFunc(err)
	}
	wg.Wait()
}
