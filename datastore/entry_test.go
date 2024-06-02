package datastore

import (
	"bufio"
	"bytes"
	"testing"
)

func TestEntry_Encode(t *testing.T) {
	e := Entry{"key", "val"}
	data := e.Encode()
	e.Decode(data)
	if e.GetLength() != 18 {
		t.Error("Incorrect length")
	}
	if e.key != "key" {
		t.Error("Incorrect key")
	}
	if e.value != "val" {
		t.Error("Incorrect value")
	}
}

func TestReadValue(t *testing.T) {
	encoder := Entry{"key", "val"}
	data := encoder.Encode()
	readData := bytes.NewReader(data)
	bReadData := bufio.NewReader(readData)
	value, err := readValue(bReadData)
	if err != nil {
		t.Fatal(err)
	}
	if value != "val" {
		t.Errorf("Wrong value: [%s]", value)
	}
}
