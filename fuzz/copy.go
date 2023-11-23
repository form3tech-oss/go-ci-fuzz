package fuzz

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func copyDirectory(dest string, src string) error {
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dest, relPath)

		if d.IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return err
			}
			return nil
		}

		if !d.Type().IsRegular() {
			return nil
		}

		return CopyFile(destPath, path, 0666)
	})

	if err != nil {
		return err
	}
	return nil
}

func CopyFile(dest, src string, perm os.FileMode) (err error) {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.OpenFile(dest, os.O_CREATE|os.O_RDWR|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("cannot create destination file")
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return err
	}
	return nil
}
