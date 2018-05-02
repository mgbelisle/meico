package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
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
	RootURL string // Relative path to the root url (e.g. "../..")
}

var (
	logPrefix     = os.Args[0] + ": "
	verboseLogger = log.New(ioutil.Discard, logPrefix, log.LstdFlags)
	errLogger     = log.New(os.Stderr, logPrefix, log.LstdFlags)
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
		verboseLogger = log.New(os.Stdout, logPrefix, log.LstdFlags)
	}

	// Build once
	build(func(err error) {
		errLogger.Panic(err)
	})

	// Serve at addr if provided
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Add(-1)
		if *addrFlag != "" {
			verboseLogger.Printf("Serving %s on %s", *outFlag, *addrFlag)
			if err := http.ListenAndServe(*addrFlag, http.FileServer(http.Dir(*outFlag))); err != nil {
				errLogger.Panic(err)
			}
		}
	}()

	// Listen for changes
	wg.Add(1)
	go func() {
		wg.Add(-1)
		prevModTime := time.Now()
		for {
			rebuild := false
			for _, path := range []string{*inFlag, *templatesFlag} {
				if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						errLogger.Panic(err)
					}
					if info.ModTime().After(prevModTime) {
						verboseLogger.Printf("Change detected in %s", path)
						rebuild = true
						prevModTime = info.ModTime()
					}
					return nil
				}); err != nil {
					errLogger.Panic(err)
				}
			}
			if rebuild {
				build(func(err error) {
					errLogger.Print(err)
				})
			}
			time.Sleep(time.Second)
		}
	}()

	wg.Wait()
}

func build(errLogFunc func(error)) {
	// Templates setup
	tmpl, err := template.ParseGlob(filepath.Join(*templatesFlag, "**"))
	if err != nil {
		errLogFunc(err)
	}

	// Render the files
	if err := os.RemoveAll(*outFlag); err != nil {
		errLogFunc(err)
	}
	wg := sync.WaitGroup{}
	if err := filepath.Walk(*inFlag, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errLogFunc(err)
		}
		rel, err := filepath.Rel(*inFlag, path)
		if err != nil {
			errLogFunc(err)
		}
		outPath := filepath.Join(*outFlag, rel)
		if info.IsDir() {
			// Make the dir
			verboseLogger.Printf("Creating %s", outPath)
			if err := os.Mkdir(outPath, info.Mode()); err != nil {
				errLogFunc(err)
			}
		} else {
			// Otherwise parse the file or copy it, whichever is appropriate.
			// Do them all in parallel
			wg.Add(1)
			go func(path string, outPath string, info os.FileInfo) {
				defer wg.Add(-1)
				outFile, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE, info.Mode())
				if err != nil {
					errLogFunc(err)
				}
				defer outFile.Close()
				rootPath, err := filepath.Rel(path, *inFlag)
				if err != nil {
					errLogFunc(err)
				}
				if filepath.Ext(path) == ".html" {
					verboseLogger.Printf("Parsing %s", path)
					tmpl2, err := tmpl.Clone()
					if err != nil {
						errLogFunc(err)
					}
					tmpl2 = template.Must(tmpl2.ParseFiles(path))
					if err := tmpl2.Execute(outFile, &TemplateData{
						RootURL: filepath.ToSlash(rootPath),
					}); err != nil {
						errLogFunc(err)
					}
				} else {
					verboseLogger.Printf("Copying %s", path)
					inFile, err := os.Open(path)
					if err != nil {
						errLogFunc(err)
					}
					defer inFile.Close()
					if _, err := io.Copy(outFile, inFile); err != nil {
						errLogFunc(err)
					}
				}
			}(path, outPath, info)
		}
		return nil
	}); err != nil {
		errLogFunc(err)
	}
	wg.Wait()
}
