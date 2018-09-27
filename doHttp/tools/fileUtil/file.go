package fileUtil

import (
	"os"
	"bufio"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreateDir(path string) error {

	pathExist, err := PathExists(path)

	if err != nil{
		return err
	}

	if !pathExist {
		err := os.MkdirAll(path, 0755)

		if err != nil {
			return err
		}
	}
	return nil
}

func CreateFile(path string, name string) (*os.File, error) {

	CreateDir(path)

	path = path + "/" + name

	pathExist, err := PathExists(path)

	if err != nil{
		return nil,err
	}

	if !pathExist {
		file, err := os.Create(path)

		if err != nil {
			return nil, err
		}

		return file, nil
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)

	if err != nil {
		return nil, err
	}

	return file, nil
}

func WriteFile(file *os.File, content string) {
	w := bufio.NewWriter(file)

	_, err := w.WriteString(content)

	defer w.Flush()

	if err != nil {
		return
	}

}
