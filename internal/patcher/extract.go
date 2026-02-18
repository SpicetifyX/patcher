package patcher

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func ExtractSPAApps(clientDir string) {
	log.Printf("Patching spotify client located at: %s\n", clientDir)

	extractTasks := []string{
		"login",
		"xpui",
	}

	var wg sync.WaitGroup

	for _, task := range extractTasks {
		wg.Add(1)

		go func() {
			defer wg.Done()

			spaPath := filepath.Join(clientDir, "Apps", fmt.Sprintf("%s.spa", task))
			destPath := filepath.Join(clientDir, "Apps", task)

			reader, err := zip.OpenReader(spaPath)
			if err != nil {
				log.Printf("Failed opening %s: %v\n", spaPath, err)
				return
			}

			defer reader.Close()

			for _, file := range reader.File {
				fullPath := filepath.Join(destPath, file.Name)

				if !strings.HasPrefix(fullPath, filepath.Clean(destPath)+string(os.PathSeparator)) {
					log.Printf("Illegal file path detected: %s\n", fullPath)
					continue
				}

				if file.FileInfo().IsDir() {
					if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
						log.Printf("Failed creating dir %s: %v\n", fullPath, err)
					}
					continue
				}

				if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
					log.Printf("Failed creating parent dir: %v\n", err)
					continue
				}

				srcFile, err := file.Open()
				if err != nil {
					log.Printf("Failed opening file in zip: %v\n", err)
					continue
				}

				dstFile, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
				if err != nil {
					log.Printf("Failed creating file %s: %v\n", fullPath, err)
					srcFile.Close()
					continue
				}

				_, err = io.Copy(dstFile, srcFile)

				srcFile.Close()
				dstFile.Close()

				if err != nil {
					log.Printf("Failed writing file %s: %v\n", fullPath, err)
				}
			}

		}()
	}

	wg.Wait()

	for _, task := range extractTasks {
		spaPath := filepath.Join(clientDir, "Apps", fmt.Sprintf("%s.spa", task))
		os.Remove(spaPath)
	}
}
