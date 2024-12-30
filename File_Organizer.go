package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sync"
)

var wg = sync.WaitGroup{}

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Sorting: %v\n", pwd)

    fileTypes := create_dir(pwd)

    moveFiles(pwd, fileTypes)
	wg.Wait()
	fmt.Println("Sorting complete.")
}

func create_dir(unsorted string) []string {
	fileTypes := make(map[string]bool)
	err := filepath.WalkDir(unsorted, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			ext := path.Ext(d.Name())
			if len(ext) > 0 && ext[0] == '.' { // Ensure we have the dot and not the empty extension
				fileTypes[ext] = true
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory %s: %v\n", unsorted, err)
		os.Exit(1)
	}

    sorted := make([]string, 0, len(fileTypes))
	for ext := range fileTypes {
		sorted = append(sorted, ext)
	}

	for _, ext := range sorted {
		err := os.MkdirAll(filepath.Join(unsorted, ext), os.ModePerm)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", ext, err)
            os.Exit(1)
		}
	}
	return sorted
}

func moveFiles(unsorted string, fileTypes []string) {
	var moveWg sync.WaitGroup

	err := filepath.WalkDir(unsorted, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		fileName := d.Name()

		for _, extFile := range fileTypes {
			if len(fileName) > len(extFile) && fileName[len(fileName)-len(extFile):] == extFile { // check if file name has correct suffix for extension
				moveWg.Add(1)
				go func() {
					defer moveWg.Done()
					dest := filepath.Join(unsorted, extFile, fileName)

					_, err := os.Stat(dest)
					if err == nil {
						fmt.Printf("File %s already exists in destination %s, not overwriting\n", filePath, dest)
						return
					}

					err = os.Rename(filePath, dest)

					if err != nil {
						fmt.Printf("Error moving %s to %s: %v\n", filePath, dest, err)
					} else {
						fmt.Printf("Moved %s to %s\n", filePath, dest)
					}
				}()
				return nil
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
        os.Exit(1)
	}
	moveWg.Wait()
}