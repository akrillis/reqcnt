package main

import (
	"bufio"
	"github.com/akrillis/reqcnt/internal/hash"
	"github.com/akrillis/reqcnt/internal/random"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
)

func Test_reader(t *testing.T) {
	tstInput := "../../test/input01.txt"
	tstDir := random.String(8)
	max := 2

	in, err := os.Open(tstInput)
	assert.Nil(t, err)
	assert.Nil(t, os.Mkdir(tstDir, dirPerm))

	assert.Nil(t, reader(bufio.NewReader(in), max, delimiter, tstDir))

	checkData := map[string]int{
		"this": 2,
		"test": 2,
		"asd":  2,
		"the":  2,
		"end":  2,
		"sad":  1,
		"is":   1,
		"my":   1,
		"only": 1,
	}

	for key, value := range checkData {
		name := tstDir + "/" + hash.Hash(key)

		data, err := os.ReadFile(name)
		assert.Nil(t, err)

		values := strings.Split(string(data)[:len(string(data))-1], "\t")

		assert.Equal(t, key, values[0])
		assert.Equal(t, strconv.Itoa(value), values[1])
	}

	assert.Nil(t, in.Close())

	//
	// remove test directory
	//

	assert.Nil(t, os.RemoveAll(tstDir))

}

func Test_worker(t *testing.T) {
	tstDir := random.String(8)

	// wrong directory
	assert.NotNil(t, worker(map[string]int{"zero": 0}, "dskjcbsdkjcb"))

	//
	// prepare test directory and data
	//

	assert.Nil(t, os.Mkdir(tstDir, dirPerm))

	cnt := 128
	testData := make(map[string]int, cnt)
	for i := 0; i < cnt; i++ {
		testData["test"+strconv.Itoa(i)] = i * 10
	}

	// success execution
	assert.Nil(t, worker(testData, tstDir))
	assert.Nil(t, worker(testData, tstDir))

	//
	// check the result
	//

	for key, value := range testData {
		name := tstDir + "/" + hash.Hash(key)

		data, err := os.ReadFile(name)
		assert.Nil(t, err)

		values := strings.Split(string(data)[:len(string(data))-1], "\t")
		found := false

		// check for value*2 because of the two executions of worker
		for i := 0; i < cnt; i++ {
			if values[0] == key && values[1] == strconv.Itoa(value*2) {
				found = true
				break
			}
		}
		assert.True(t, found)
	}

	//
	// remove test directory
	//

	assert.Nil(t, os.RemoveAll(tstDir))
}

func Test_writer(t *testing.T) {
	tstDir := random.String(8)
	tstOut := "out.txt"

	// wrong directory
	assert.NotNil(t, writer("bskhcbskchbsk", 4, tstOut))

	//
	// prepare test directory
	//

	assert.Nil(t, os.Mkdir(tstDir, dirPerm))

	// wrong output file
	assert.NotNil(t, writer(tstDir, 4, "hsvasjhavs/"+tstOut))

	//
	// prepare test files
	//

	testData := []struct{ name, content string }{}
	cnt := 128
	for i := 0; i < cnt; i++ {
		testData = append(testData, struct{ name, content string }{tstDir + "/test" + strconv.Itoa(i), "test" + strconv.Itoa(i) + "\t" + strconv.Itoa(i*10) + "\n"})
	}

	for _, td := range testData {
		file, err := os.Create(td.name)
		assert.Nil(t, err)
		_, err = file.WriteString(td.content)
		assert.Nil(t, err)
		assert.Nil(t, file.Close())
	}

	// success execution
	assert.Nil(t, writer(tstDir, 4, tstOut))

	//
	// check the result
	//

	out, err := os.OpenFile(tstOut, os.O_RDONLY, filePerm)
	assert.Nil(t, err)
	reader := bufio.NewReader(out)
	for {
		line, err := reader.ReadString(delimiter)

		if err == io.EOF {
			break
		} else {
			assert.Nil(t, err)
		}

		found := false
		for _, td := range testData {
			if td.content == line {
				found = true
				break
			}
		}
		assert.True(t, found)
	}

	assert.Nil(t, out.Close())

	//
	// remove test infrastructure
	//

	assert.Nil(t, os.RemoveAll(tstDir))
	assert.Nil(t, os.Remove(tstOut))
}
