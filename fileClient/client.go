package fileClient

import (
	"errors"
	"fmt"
	"os"
)

type FileClient struct {
	File     *os.File
	FileName string
}

func NewFileClient(fileName string) *FileClient {
	return &FileClient{
		FileName: fileName,
	}
}

// InitFileConn: checks if file already exists, opens connection to file.
func (fc *FileClient) InitFileConn() error {
	var f *os.File
	exists, err := fc.valueFileExists()
	if err != nil {
		return err
	}
	if !exists {
		f, err = os.Create(fc.FileName)
		if err != nil {
			return err
		}

	} else {
		f, err = os.OpenFile(fc.FileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
	}
	fc.File = f
	return nil
}

// valueFileExists: looks for file value by name.
func (fc *FileClient) valueFileExists() (bool, error) {
	_, err := os.Stat(fc.FileName)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (fc *FileClient) WriteToFile(str string) {
	fmt.Fprintln(fc.File, str)
}
