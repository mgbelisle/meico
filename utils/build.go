package main

import (
	"errors"
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
	templatesFlag = flag.String("templates", "templates/base.html templates/**", "String separated list of templates. The first one is the base template")
	verboseFlag   = flag.Bool("verbose", false, "Verbose output")
	addrFlag      = flag.String("addr", "", "Address to serve output dir, if provided")
)

type TemplateData struct {
	RootURL string // Relative path to the root url, relative to --in (e.g. "../..")
	Path    string // Relative path of the file being parsed, relative to --in (e.g. "contact/index.html")
}

var TemplateFuncs = template.FuncMap{
	// {{ dict "fooKey" "fooVal" "barKey" "barValue" }}
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, errors.New("dict must have an even number of args")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, errors.New("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	},
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

	wg := sync.WaitGroup{}
	if *addrFlag != "" {
		// Serve at addr if provided
		wg.Add(1)
		go func() {
			defer wg.Add(-1)
			verboseLogger.Printf("Serving %s on %s", *outFlag, *addrFlag)
			if err := http.ListenAndServe(*addrFlag, http.FileServer(http.Dir(*outFlag))); err != nil {
				errLogger.Panic(err)
			}
		}()

		// Listen for changes
		wg.Add(1)
		go func() {
			defer wg.Add(-1)
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
	}

	wg.Wait()
}

func build(errLogFunc func(error)) {
	var tmpl *template.Template
	// Templates setup
	for _, glob in strings.Fields(*templatesFlag) {
		paths, err := filepath.Glob(glob)
		if err != nil {
			errLogFunc(err)
			return
		}
		if tmpl == nil && 0 < len(paths) {
			tmpl = template.New(paths[0])
		} else {
			for _, path := range paths {
				tmpl = tmpl.New(path)
			}
		}
	}
	if tmpl == nil {
		errLogFunc(errors.New("No templates found"))
		return
	}
	tmpl = tmpl.Funcs(TemplateFuncs)
	var err error
	tmpl, err = 
	if err != nil {
		errLogFunc(err)
		return
	}

	// Render the files
	if err := os.RemoveAll(*outFlag); err != nil {
		errLogFunc(err)
		return
	}
	wg := sync.WaitGroup{}
	if err := filepath.Walk(*inFlag, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(*inFlag, path)
		if err != nil {
			return err
		}
		outPath := filepath.Join(*outFlag, relPath)
		if info.IsDir() {
			// Make the dir
			verboseLogger.Printf("Creating %s", outPath)
			if err := os.Mkdir(outPath, info.Mode()); err != nil {
				return err
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
					return
				}
				defer outFile.Close()
				rootPath, err := filepath.Rel(filepath.Dir(path), *inFlag)
				if err != nil {
					errLogFunc(err)
					return
				}
				if filepath.Ext(path) == ".html" {
					verboseLogger.Printf("Parsing %s", path)
					tmpl2, err := tmpl.Clone()
					if err != nil {
						errLogFunc(err)
						return
					}
					tmpl2, err = tmpl2.ParseFiles(path)
					if err != nil {
						errLogFunc(err)
						return
					}
					if err := tmpl2.Execute(outFile, &TemplateData{
						RootURL: filepath.ToSlash(rootPath),
						Path:    relPath,
					}); err != nil {
						errLogFunc(err)
						return
					}
				} else {
					verboseLogger.Printf("Copying %s", path)
					inFile, err := os.Open(path)
					if err != nil {
						errLogFunc(err)
						return
					}
					defer inFile.Close()
					if _, err := io.Copy(outFile, inFile); err != nil {
						errLogFunc(err)
						return
					}
				}
			}(path, outPath, info)
		}
		return nil
	}); err != nil {
		errLogFunc(err)
		return
	}
	wg.Wait()
}
