package database

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"testing"
)

var M map[string][]byte
var A []map[string][]byte

func TestMain(m *testing.M) {
	M = createMap(0)
	A = createArray(10)

	result := m.Run()

	os.Exit(result)
}

func TestWriteRead(t *testing.T) {
	setUp()

	err := Write("parent", "child", M)
	ok(t, err)

	m, err := Read("parent", "child")
	ok(t, err)
	equals(t, M, m)

	tearDown()
}

func TestWritesReads(t *testing.T) {
	setUp()

	err := Writes("parent", "key", A)
	ok(t, err)

	a, err := Reads("parent")
	ok(t, err)
	equals(t, A, a)

	tearDown()
}

func TestWriteFieldReadField(t *testing.T) {
	setUp()

	err := Write("parent", "child", M)
	ok(t, err)

	err = WriteField("parent", "child", "key", []byte("changed"))
	ok(t, err)

	f, err := ReadField("parent", "child", "key")
	ok(t, err)
	equals(t, []byte("changed"), f)

	tearDown()
}

func TestWritesFieldReadsField(t *testing.T) {
	setUp()

	err := Write("parent", "child1", M)
	ok(t, err)
	err = Write("parent", "child2", M)
	ok(t, err)

	err = WritesField("parent", "key", []byte("changed"))
	ok(t, err)

	f, err := ReadsField("parent", "key")
	ok(t, err)
	equals(t, []byte("changed"), f["child1"])
	equals(t, []byte("changed"), f["child2"])

	tearDown()
}

func TestWriteMappingReadMapping(t *testing.T) {
	setUp()

	err := WriteMapping("parent", M)
	ok(t, err)

	m, err := ReadMapping("parent")
	ok(t, err)
	equals(t, M, m)

	tearDown()
}

func TestDelete(t *testing.T) {
	setUp()

	err := Write("parent", "child", M)
	ok(t, err)

	err = Delete("parent", "child")
	ok(t, err)

	m, err := Read("parent", "child")
	equals(t, "Bucket child not found", err.Error())
	equals(t, 0, len(m))

	tearDown()
}

func TestDeletes(t *testing.T) {
	setUp()

	err := Write("parent", "child1", M)
	ok(t, err)

	err = Write("parent", "child2", M)
	ok(t, err)

	err = Deletes("parent")
	ok(t, err)

	m, err := Read("parent", "child")
	equals(t, "Bucket parent not found", err.Error())
	equals(t, 0, len(m))

	tearDown()
}

func TestDeleteMapping(t *testing.T) {
	setUp()

	err := WriteMapping("parent", M)
	ok(t, err)

	err = DeleteMapping("parent")
	ok(t, err)

	m, err := ReadMapping("parent")
	equals(t, "Bucket parent not found", err.Error())
	equals(t, 0, len(m))

	tearDown()
}

func createMap(key int) map[string][]byte {
	type enum string
	const (
		test enum = "test"
	)

	m := make(map[string][]byte)

	m["key"] = []byte(strconv.FormatInt((int64)(key), 10))
	m["key1"] = []byte("value")                                //string
	m["key2"] = []byte(strconv.FormatInt(1, 10))               //int
	m["key3"] = []byte(strconv.FormatFloat(1.11, 'f', -1, 32)) //float32
	m["key4"] = []byte(strconv.FormatFloat(1.11, 'f', -1, 64)) //float64
	m["key5"] = []byte(strconv.FormatBool(true))               //bool
	m["key6"] = []byte(test)                                   //type

	return m
}

func createArray(n int) []map[string][]byte {
	a := make([]map[string][]byte, 0, n)

	for i := 0; i < n; i++ {
		a = append(a, createMap(i))
	}

	return a
}

func setUp() {
	initialise(".", "test.db")
}

func tearDown() {
	os.Remove("test.db")
}

func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
