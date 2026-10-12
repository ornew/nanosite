// Command nanosite builds static documentation from an arbitrary repository.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	repo := flag.String("repo", ".", "input repository (read only)")
	config := flag.String("config", "site/config.json", "JSON configuration file")
	out := flag.String("out", "site/dist", "generated output directory")
	serve := flag.String("serve", "", "serve after building, e.g. localhost:8080")
	check := flag.Bool("check", false, "validate without writing output")
	drafts := flag.Bool("drafts", false, "include draft pages in a local preview")
	base := flag.String("base-path", "", "override URL deployment prefix, e.g. /project/")
	flag.Parse()
	if flag.NArg() != 0 {
		log.Fatal("unexpected positional arguments")
	}
	c, err := readConfig(*config)
	if err != nil {
		log.Fatal(err)
	}
	if *drafts {
		c.Drafts = true
	}
	if *base != "" {
		if !strings.HasPrefix(*base, "/") || strings.ContainsAny(*base, "?#\\") || strings.Contains(*base, "..") {
			log.Fatal("invalid -base-path")
		}
		c.BasePath = "/" + strings.Trim(*base, "/")
		if c.BasePath != "/" {
			c.BasePath += "/"
		}
	}
	result, err := build(*repo, *out, c, *check)
	if err != nil {
		log.Fatal(err)
	}
	if *check {
		fmt.Printf("validated %d pages (no output written)\n", len(result.Pages))
		return
	}
	fmt.Printf("built %d pages in %s\n", len(result.Pages), result.Out)
	if *serve != "" {
		handler := http.FileServer(http.Dir(result.Out))
		mux := http.NewServeMux()
		if c.BasePath == "/" {
			mux.Handle("/", handler)
		} else {
			mux.Handle(c.BasePath, http.StripPrefix(strings.TrimSuffix(c.BasePath, "/"), handler))
		}
		// Preview never lists files outside the generated output directory.
		log.Printf("preview: http://%s%s (rebuild after edits)", *serve, c.BasePath)
		if err := http.ListenAndServe(*serve, mux); err != nil {
			log.Fatal(err)
		}
	}
}

func canonicalPath(p string) (string, error) {
	a, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	if _, err := os.Lstat(a); os.IsNotExist(err) {
		parent, err := canonicalPath(filepath.Dir(a))
		if err != nil {
			return "", err
		}
		return filepath.Join(parent, filepath.Base(a)), nil
	}
	return filepath.EvalSymlinks(a)
}
