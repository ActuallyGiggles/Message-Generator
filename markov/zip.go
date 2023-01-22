package markov

import (
	"archive/zip"
	"io"
	"os"
	"strings"
)

func zipChains() {
	busy.Lock()
	defer duration(track("zip duration"))

	debugLog("creating zip archive...")
	archive, err := os.Create("markov-chains.zip")
	if err != nil {
		panic(err)
	}
	defer archive.Close()
	zipWriter := zip.NewWriter(archive)

	if err := addDirectoryToZip(zipWriter, "./markov-chains/"); err != nil {
		panic(err)
	}

	debugLog("closing zip archive...")
	zipWriter.Close()
	busy.Unlock()
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
		debugLog("zipping directory", filePath, "to archive...")
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
