package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/akrillis/reqcnt/internal/hash"
	"github.com/akrillis/reqcnt/internal/random"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	defaultInputPath   = "input.txt"
	defaultOutputPath  = "output.txt"
	defaultRequestsQty = 4

	tmpDirLength = 16
	dirPerm      = 0755
	filePerm     = 0666
	delimiter    = '\n'
	content      = "%s\t%d\n"

	errOpenFile      = "could not open file %s: %v\n"
	errCreateDir     = "could not create directory %s: %v\n"
	errOpenDir       = "could not open directory %s: %v\n"
	errReadDir       = "could not read directory %s: %v\n"
	errCouldNotRead  = "could not read from file %s: %v\n"
	errCouldNotWrite = "could not write to file %s: %v\n"

	errCreateFileForRequest = "could not create temporary file for request %s: %v\n"
	errOpenFileForRequest   = "could not open temporary file for request %s: %v\n"
	errFileStatForRequest   = "could not get file stat for request %s: %v\n"
	errReadFileForRequest   = "could not read from file for request %s: %v\n"
	errParseValueForRequest = "could not parse value for request %s: %v\n"
	errWriteFileForRequest  = "could not write to file for request %s: %v\n"
	errCloseFileForRequest  = "could not close file for request %s: %v\n"

	errInputParse  = "could not parse input data: %v\n"
	errWriteOutput = "could not write output data: %v\n"
)

var (
	input  = flag.String("input", defaultInputPath, "input file with requests")
	output = flag.String("output", defaultOutputPath, "output file with results")
	qty    = flag.Int("qty", defaultRequestsQty, "maximum quantity of requests in memory (greater than 0)")
)

func main() {
	flag.Parse()

	if *qty < 1 {
		log.Fatalf("qty must be greater than 0")
	}

	in, err := os.Open(*input)
	if err != nil {
		log.Fatalf(errOpenFile, *input, err)
	}
	defer func() {
		_ = in.Close()
	}()

	tempDir := random.String(tmpDirLength)
	if err := os.Mkdir(tempDir, dirPerm); err != nil {
		log.Fatalf(errCreateDir, tempDir, err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	if err := reader(bufio.NewReader(in), *qty, delimiter, tempDir); err != nil {
		log.Fatalf(errInputParse, err)
	}

	if err := writer(tempDir, *qty, *output); err != nil {
		log.Fatalf(errWriteOutput, err)
	}
}

// Читает и обрабатывает входной файл
//
//  1. Читает содержимое входного файла построчно
//  2. Считает количество запросов (пустые строки игнорируются)
//  3. Если количество запросов достигло максимального, то вызывает функцию worker и очищает буфер запросов
//  4. Повторяет 1-3 пока не достигнет конца входного файла
func reader(reader *bufio.Reader, max int, separator byte, tmpDir string) error {
	reqCounter := make(map[string]int, max)

	for {
		line, err := reader.ReadString(separator)
		if err != nil && err != io.EOF {
			log.Fatalf(errCouldNotRead, *input, err)
		}

		if line[len(line)-1] == separator {
			line = line[:len(line)-1]
		}

		if len(line) == 0 {
			continue
		}

		_, ok := reqCounter[line]
		if ok {
			reqCounter[line]++
		} else {
			reqCounter[line] = 1
		}

		if err == io.EOF {
			break
		}

		if len(reqCounter) == max {
			if err := worker(reqCounter, tmpDir); err != nil {
				return err
			}
			reqCounter = make(map[string]int, max)
		}
	}

	if err := worker(reqCounter, tmpDir); err != nil {
		return err
	}

	reqCounter = make(map[string]int)
	return nil
}

// Обрабатывает буфер запросов
//
//  1. Принимает на вход буфер запросов в виде map[string]int, где ключ - запрос, а значение - количество запросов
//  2. Проверяет наличие файла с именем, полученным из хеша ключа
//  3. Если файла нет, то создает его и записывает в него значение
//  4. Если файл есть, то открывает его и увеличивает счетчик на значение из буфера
//  5. Повторяет 2-4 пока не обработает все запросы
func worker(input map[string]int, tmpDir string) error {
	for key, value := range input {
		name := tmpDir + "/" + hash.Hash(key)

		var file *os.File
		_, err := os.Stat(name)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf(errFileStatForRequest, key, err)
		}

		switch os.IsNotExist(err) {
		case true:
			file, err = os.Create(name)
			if err != nil {
				return fmt.Errorf(errCreateFileForRequest, key, err)
			}

		case false:
			file, err = os.OpenFile(name, os.O_RDWR, filePerm)
			if err != nil {
				return fmt.Errorf(errOpenFileForRequest, key, err)
			}

			nr := bufio.NewReader(file)

			line, err := nr.ReadString('\n')
			if err != nil {
				return fmt.Errorf(errReadFileForRequest, key, err)
			}

			values := strings.Split(line[:len(line)-1], "\t")
			cnt, err := strconv.Atoi(values[1])
			if err != nil {
				return fmt.Errorf(errParseValueForRequest, key, err)
			}

			value += cnt
		}

		_, err = file.WriteAt([]byte(fmt.Sprintf(content, key, value)), 0)
		if err != nil {
			return fmt.Errorf(errWriteFileForRequest, key, err)
		}

		if err := file.Close(); err != nil {
			return fmt.Errorf(errCloseFileForRequest, key, err)
		}
	}

	return nil
}

// Формирует выходной файл
//
//  1. Открывает директорию с временными файлами
//  2. Читает содержимое директории итеративно с ограничением на количество файлов равному максимальному количеству запросов
//  3. Открывает каждый файл и читает его содержимое
//  4. Записывает содержимое в выходной файл
//  5. Повторяет 2-4 пока не прочитает все файлы
func writer(dir string, max int, outFile string) error {
	d, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf(errOpenDir, dir, err)
	}
	defer func() {
		_ = d.Close()
	}()

	out, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY, filePerm)
	if err != nil {
		return fmt.Errorf(errOpenFile, outFile, err)
	}
	defer func() {
		_ = out.Close()
	}()

	for {
		files, err := d.Readdir(max)

		if err != nil && err != io.EOF {
			return fmt.Errorf(errReadDir, dir, err)
		}

		for _, file := range files {
			name := dir + "/" + file.Name()
			data, err := os.ReadFile(name)
			if err != nil {
				return fmt.Errorf(errCouldNotRead, name, err)
			}

			if _, err := out.Write(data); err != nil {
				return fmt.Errorf(errCouldNotWrite, outFile, err)
			}
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}
