package markov

import (
	"archive/zip"
	"io"
	"os"
	"strings"
	"time"
)

func zipChains() {
	if !instructions.ShouldZip {
		return
	}
	busy.Lock()
	defer busy.Unlock()
	defer duration(track("zipping duration"))

	defaultPath := "./markov-chains.zip"
	newPath := "./markov-chains_new.zip"

	archive, err := os.Create(newPath)
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	if err := addDirectoryToZip(zipWriter, "./markov-chains/"); err != nil {
		panic(err)
	}

	removeAndRename(defaultPath, newPath)

	zipWriter.Close()
	stats.NextZipTime = time.Now().Add(zipInterval)
}

func addDirectoryToZip(zipWriter *zip.Writer, path string) error {
	if !strings.HasPrefix(path, "./") {
		path = "./" + path
	}
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		filePath := path + file.Name()

		if file.IsDir() {
			if err := addDirectoryToZip(zipWriter, filePath); err != nil {
				return err
			}
			continue
		}

		exists, w := doesWorkerExist(file.Name())
		if exists {
			w.ChainMx.Lock()
			defer w.ChainMx.Unlock()
		}

		f2, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f2.Close()

		filePath = strings.TrimPrefix(filePath, "./")
		w2, err := zipWriter.Create(filePath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(w2, f2); err != nil {
			return err
		}
	}

	return nil
}
