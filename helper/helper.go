package helper

import (
	"archive/tar"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func GetApplication(applicationsPath, applicationUUID string) (string, string, error) {
	commit, err := getCommit(applicationsPath, applicationUUID)
	if err != nil {
		log.Printf("Failed to get commit. Error: %s\r\n", err)
		return "", "", err
	}

	basePath := filepath.Join(applicationsPath, applicationUUID, commit)
	tarball := filepath.Join(basePath, "binary.tar")
	application := filepath.Join(basePath, "application.zip")

	_, err = os.Stat(application)
	if err != nil {
		err = untar(tarball, basePath)
		if err != nil {
			log.Printf("Failed to extract application. Error: %s\r\n", err)
			return "", "", err
		}
	}

	return application, commit, err
}

func untar(tarball, target string) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()

	tarReader := tar.NewReader(reader)
	for header, err := tarReader.Next(); err != io.EOF; header, err = tarReader.Next() {
		if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}

func getCommit(applicationsPath, applicationUUID string) (string, error) {
	path := filepath.Join(applicationsPath, applicationUUID)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return "", err
	} else if len(files) != 1 {
		return "", errors.New("Expected single commit")
	}
	return files[0].Name(), nil
}
