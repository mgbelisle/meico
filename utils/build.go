package main

import (
	"encoding/json"
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
	"strings"
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
	dataFlag      = flag.String("data", "data", "Data dir (for json data)")
	templatesFlag = flag.String("templates", "templates/base.html templates/**.html", "String separated list of template globs. The first one is the base template (required)")
	verboseFlag   = flag.Bool("verbose", false, "Verbose output")
	addrFlag      = flag.String("addr", "", "Address to serve output dir, if provided")
)

type TemplateData struct {
	RootURL string // Relative path to the root url, relative to --in (e.g. "../..")
	Path    string // Relative path of the file being parsed, relative to --in (e.g. "contact/index.html")
}

var TemplateFuncs = template.FuncMap{
	"json": func(file string) (interface{}, error) {
		data, err := ioutil.ReadFile(filepath.Join(*dataFlag, file))
		if err != nil {
			return nil, err
		}
		var obj interface{}
		return obj, json.Unmarshal(data, &obj)
	},
	"string": func(value interface{}) (string, error) {
		if value, ok := value.(string); ok {
			return value, nil
		}
		return "", fmt.Errorf("Not a string: %v (%T)", value, value)
	},
	"active": func(url string, templateData *TemplateData) bool {
		fmt.Printf("%s %v\n", url, templateData)
		return false
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
				for _, glob := range append([]string{
					filepath.Join(*inFlag, "**"),
					filepath.Join(*dataFlag, "**"),
				}, strings.Fields(*templatesFlag)...) {
					paths, err := filepath.Glob(glob)
					if err != nil {
						errLogger.Print(err)
						break
					}
					for _, path := range paths {
						info, err := os.Stat(path)
						if err != nil {
							errLogger.Print(err)
							break
						}
						if info.ModTime().After(prevModTime) {
							verboseLogger.Printf("Change detected in %s", path)
							rebuild = true
							prevModTime = info.ModTime()
						}
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
	// Templates setup
	templatesFields := strings.Fields(*templatesFlag)
	if len(templatesFields) < 1 {
		errLogFunc(errors.New("--templates requires at least the base template"))
		return
	}
	basePaths, err := filepath.Glob(templatesFields[0])
	if err != nil {
		errLogFunc(err)
		return
	}
	if l := len(basePaths); l != 1 {
		errLogFunc(fmt.Errorf("%d base templates exist, exactly 1 required", l))
		return
	}
	tmpl, err := template.New(filepath.Base(basePaths[0])).Funcs(TemplateFuncs).ParseFiles(basePaths[0])
	if err != nil {
		errLogFunc(err)
		return
	}
	verboseLogger.Printf("Parsed base template: %s", basePaths[0])
	for _, glob := range templatesFields[1:] {
		tmpl, err = tmpl.ParseGlob(glob)
		if err != nil {
			errLogFunc(err)
			return
		}
		verboseLogger.Printf("Parsed templates: %s", glob)
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
			verboseLogger.Printf("Creating dir: %s", outPath)
			if err := os.Mkdir(outPath, info.Mode()); err != nil {
				return err
			}
		} else {
			// Otherwise execute the template or copy the file, whichever is appropriate.
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
				if tmpl != nil && filepath.Ext(path) == ".html" {
					verboseLogger.Printf("Executing template: %s", path)
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
					verboseLogger.Printf("Copying file: %s", path)
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
