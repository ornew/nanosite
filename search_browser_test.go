package main

import (
	"os/exec"
	"testing"
)

func TestSearchBrowserHeadingURLs(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("Node.js is not installed")
	}
	output, err := exec.Command(node, "testdata/search_url_test.cjs", "static/site.js").CombinedOutput()
	if err != nil {
		t.Fatalf("browser search URL test: %v\n%s", err, output)
	}
}
