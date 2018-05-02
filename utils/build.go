package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var usagePrefix = fmt.Sprintf(`Builds a static site using the html/template package

Usage: %s [OPTIONS]

OPTIONS:
`, os.Args[0])

var (
	inFlag        = flag.String("in", "src", "Input dir")
	outFlag       = flag.String("out", "www", "Output dir")
	templatesFlag = flag.String("templates", "templates", "Templates dir")
	verboseFlag   = flag.Bool("verbose", false, "Verbose output")
)

func main() {
	// Flag setup
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usagePrefix)
		flag.PrintDefaults()
	}
	flag.Parse()

	// Logger setup
	verboseLogger := log.New(ioutil.Discard, os.Args[0], log.LstdFlags)
	if *verboseFlag {
		verboseLogger = log.New(os.Stdout, os.Args[0], log.LstdFlags)
	}

	// Templates setup
	tmpl := template.Must(template.ParseGlob(filepath.Join(*templatesFlag, "**")))

	// Render the files
	os.RemoveAll(*outFlag)
	if err := filepath.Walk(*inFlag, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Panic(err)
		}
		rel, err := filepath.Rel(*inFlag, path)
		if err != nil {
			log.Panic(err)
		}
		outPath := filepath.Join(*outFlag, rel)
		if err := os.MkdirAll(filepath.Dir(outPath), 0700); err != nil {
			log.Panic(err)
		}
		outFile, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE, info.Mode())
		if err != nil {
			log.Panic(err)
		}
		defer outFile.Close()
		if !info.IsDir() && filepath.Ext(path) == ".html" {
			verboseLogger.Printf("Parsing %s", path)
			tmpl2, err := tmpl.Clone()
			if err != nil {
				log.Panic(err)
			}
			tmpl2 = template.Must(tmpl2.ParseFiles(path))
			if err := tmpl2.Execute(outFile, nil); err != nil {
				log.Panic(err)
			}
		} else {
			verboseLogger.Printf("Copying %s", path)
			inFile, err := os.Open(path)
			if err != nil {
				log.Panic(err)
			}
			defer inFile.Close()
			if _, err := io.Copy(outFile, inFile); err != nil {
				log.Panic(err)
			}
		}
		return nil
	}); err != nil {
		log.Panic(err)
	}
}
