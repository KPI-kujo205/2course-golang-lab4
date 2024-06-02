package datastore

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestDb_Put(t *testing.T) {
	saveDirectory, err := ioutil.TempDir("", "testDir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(saveDirectory)

	dataBase, err := NewDb(saveDirectory, 45)
	if err != nil {
		t.Fatal(err)
	}
	defer dataBase.Close()

	pairs := [][]string{
		{"1", "v1"},
		{"2", "v2"},
		{"3", "v3"},
	}
	finalPath := filepath.Join(saveDirectory, outFileName+"0")
	outFile, err := os.Open(finalPath)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("check put and get methods", func(t *testing.T) {
		for _, pair := range pairs {
			err := dataBase.Put(pair[0], pair[1])
			if err != nil {
				t.Errorf("Unable to place %s: %s.", pair[0], err)
			}
			actual, err := dataBase.Get(pair[0])
			if err != nil {
				t.Errorf("Unable to retrieve %s: %s", pair[0], err)
			}
			if actual != pair[1] {
				t.Errorf("Invalid value returned. Expected: %s, Actual: %s.", pair[1], actual)
			}
		}
	})

	outInfo, err := outFile.Stat()
	if err != nil {
		t.Fatal(err)
	}
	expectedStateSize := outInfo.Size()

	t.Run("check increase file size", func(t *testing.T) {
		for _, pair := range pairs {
			err := dataBase.Put(pair[0], pair[1])
			if err != nil {
				t.Errorf("Unable to place %s: %s.", pair[0], err)
			}
		}
		t.Log(dataBase)
		outInfo, err := outFile.Stat()
		actualStateSize := outInfo.Size()
		if err != nil {
			t.Fatal(err)
		}
		if expectedStateSize != actualStateSize {
			t.Errorf("Size mismatch: Expected: %d, Actual: %d.", expectedStateSize, actualStateSize)
		}
	})

	t.Run("check creation new process test", func(t *testing.T) {
		if err := dataBase.Close(); err != nil {
			t.Fatal(err)
		}
		dataBase, err = NewDb(saveDirectory, 45)
		if err != nil {
			t.Fatal(err)
		}

		for _, pair := range pairs {
			actual, err := dataBase.Get(pair[0])
			if err != nil {
				t.Errorf("Unable to place %s: %s.", pair[1], err)
			}
			expected := pair[1]
			if actual != expected {
				t.Errorf("Invalid value returned. Expected: %s, Actual: %s.", expected, actual)
			}
		}
	})
}
