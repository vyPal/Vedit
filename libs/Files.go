package libs

import (
	"errors"
	"os"
)

type File struct {
	File     *os.File
	Name     string
	Path     string
	Contents []rune
}

type Directory struct {
	Name        string
	Path        string
	Files       []File
	Directories []Directory
}

func NewFile(name, path, contents string) File {
	return File{
		Name:     name,
		Path:     path,
		Contents: []rune(contents),
	}
}

func NewDirectory(name, path string) Directory {
	return Directory{
		Name:        name,
		Path:        path,
		Files:       []File{},
		Directories: []Directory{},
	}
}

func (f *File) Touch() error {
	if f.Path == "" {
		return errors.New("File path is empty")
	} else if f.Name == "" {
		return errors.New("File name is empty")
	}

	if _, err := os.Stat(f.Path + f.Name); err != nil && os.IsNotExist(err) {
		file, err := os.Create(f.Path + f.Name)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	return nil
}

func (f *File) Load() error {
	if f.Path == "" {
		return errors.New("File path is empty")
	} else if f.Name == "" {
		return errors.New("File name is empty")
	}

	file, err := os.Open(f.Path + f.Name)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	data := make([]byte, stat.Size())
	_, err = file.Read(data)
	if err != nil {
		return err
	}

	f.Contents = []rune(string(data))
	return nil
}

func (f *File) Save() error {
	if f.Path == "" {
		return errors.New("File path is empty")
	} else if f.Name == "" {
		return errors.New("File name is empty")
	}

	file, err := os.Create(f.Path + f.Name)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(string(f.Contents))
	if err != nil {
		return err
	}
	return nil
}
