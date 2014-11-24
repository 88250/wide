package util

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type myzip struct{}

// Zip utilities.
var Zip = myzip{}

type ZipFile struct {
	zipFile *os.File
	writer  *zip.Writer
}

func (*myzip) Create(filename string) (*ZipFile, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return &ZipFile{zipFile: file, writer: zip.NewWriter(file)}, nil
}

func (z *ZipFile) Close() error {
	err := z.zipFile.Close() // close the underlying writer
	if nil != err {
		return err
	}

	return z.writer.Close()
}

func (z *ZipFile) AddEntryN(dir string, names ...string) error {
	for _, name := range names {
		zipPath := filepath.Join(dir, name)
		err := z.AddEntry(zipPath, name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (z *ZipFile) AddEntry(dir, name string) error {
	entry, err := z.writer.Create(dir)
	if err != nil {
		return err
	}

	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(entry, file)

	return err
}

func (z *ZipFile) AddDirectoryN(dir string, names ...string) error {
	for _, name := range names {
		err := z.AddDirectory(dir, name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (z *ZipFile) AddDirectory(dir, dirName string) error {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		return err
	}

	for _, file := range files {
		localPath := filepath.Join(dirName, file.Name())
		zipPath := filepath.Join(dir, file.Name())

		err = nil
		if file.IsDir() {
			z.AddEntry(dir, dirName)

			err = z.AddDirectory(zipPath, localPath)
		} else {
			err = z.AddEntry(zipPath, localPath)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
