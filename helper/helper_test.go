package helper

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestGetApplication(t *testing.T) {
	applicationsPath := filepath.Join(os.Getenv("GOPATH"), "src/edge-node-manager/helper")
	applicationUUID := "test_app"
	result := filepath.Join(applicationsPath, applicationUUID, "test_commit", "application.zip")

	application, commit, err := GetApplication(applicationsPath, applicationUUID)
	if err != nil {
		t.Error(err)
	} else if application != result {
		t.Error("expected" + result)
	} else if commit != "test_commit" {
		t.Error("expected" + result)
	}
}

func TestUntar(t *testing.T) {
	basePath := filepath.Join(os.Getenv("GOPATH"), "src/edge-node-manager/helper/test_app/test_commit")
	sourcePath := filepath.Join(basePath, "binary.tar")
	targetPath := filepath.Join(basePath, "/application.zip")

	os.Remove(targetPath)

	err := untar(sourcePath, basePath)
	if err != nil {
		t.Error(err)
	}

	_, err = os.Stat(targetPath)
	if err != nil {
		t.Error(err)
	}

	os.Remove(targetPath)
}

func TestGetCommit(t *testing.T) {
	applicationsPath := filepath.Join(os.Getenv("GOPATH"), "src/edge-node-manager/helper")
	applicationUUID := "test_app"

	commit, err := getCommit(applicationsPath, applicationUUID)
	if err != nil {
		t.Error(err)
	} else if commit != "test_commit" {
		t.Error("Incorrect commit returned")
	}
}
