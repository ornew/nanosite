package sitegen

import (
	"os"
	"path/filepath"
)

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
