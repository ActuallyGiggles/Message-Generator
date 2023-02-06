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
	defer duration(track("zipping duration"))

	archive, err := os.OpenFile("markov-chains.zip", os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	zipWriter := zip.NewWriter(archive)

	if err := addDirectoryToZip(zipWriter, "./markov-chains/"); err != nil {
		panic(err)
	}

	zipWriter.Close()
	archive.Close()
	busy.Unlock()
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
