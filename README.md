# Request counter

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Description

Утилита для подсчета уникальных запросов в файле с ограничением по количеству запросов в памяти.

## Usage

```text
Usage of ./requestCounter:
  -input string
        input file with requests (default "input.txt")
  -output string
        output file with results (default "output.txt")
  -qty int
        maximum quantity of requests in memory (default 4)
```

## Test coverage

```text
ok      github.com/akrillis/reqcnt/cmd/requestCounter   0.686s  coverage: 66.3% of statements
ok      github.com/akrillis/reqcnt/internal/hash        0.447s  coverage: 100.0% of statements
ok      github.com/akrillis/reqcnt/internal/random      0.242s  coverage: 100.0% of statements
```

## Example

Из директории cmd/requestCounter:

```text
> cat ../../test/input01.txt 

> go build

> ./requestCounter -qty 3 -input ../../test/input01.txt -output out.txt

> cat out.txt 
```