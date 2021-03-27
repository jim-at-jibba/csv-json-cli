package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type inputFile struct {
	filepath  string
	separator string
	pretty    bool
}

func main() {
	run()
}

func run() error {
	fmt.Println("Hello")
	input, err := getFileData()

	if err != nil {
		return err
	}

	fmt.Printf("Input %v: ", input)

	return nil
}

func getFileData() (inputFile, error) {
	// we need to validate that we are getting the
	// correct number of arguments
	if len(os.Args) < 2 {
		return inputFile{}, errors.New("a filepath argument is required")
	}

	// defining the options flags
	separator := flag.String("separator", "comma", "Column separator")
	pretty := flag.Bool("pretty", false, "Generate pretty JSON")

	flag.Parse()

	fileLocation := flag.Arg(0)

	// Validating if we received a comma or semicolon
	if !(*separator == "comma" || *separator == "semicolon") {
		return inputFile{}, errors.New("only comma or semicolon separators are allowed")
	}

	return inputFile{fileLocation, *separator, *pretty}, nil
}

func checkIfValidFile(filename string) (bool, error) {
	// Check if entered file is CSV
	if fileExtension := filepath.Ext(filename); fileExtension != ".csv" {
		return false, fmt.Errorf("file %s is not a CSV", filename)
	}

	// check if filepath entered belongs to an existing file
	if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
		return false, fmt.Errorf("file %s does not exist", filename)
	}

	return true, nil
}

func processCsvFile(fileData inputFile, writerChannel chan<- map[string]string) {
	// open file for reading
	file, err := os.Open(fileData.filepath)

	// check errors
	check(err)

	// Close file when finished
	defer file.Close()

	// define headers and line
	var headers, line []string

	reader := csv.NewReader(file)

	// default character separator is comma, check to semicolon is that flag is set
	if fileData.separator == "semicolon" {
		reader.Comma = ';'
	}

	// read first line to get headers
	headers, err = reader.Read()
	check(err)

	for {
		line, err = reader.Read()
		// if we get to the EOF close channel
		if err == io.EOF {
			close(writerChannel)
			break
		} else if err != nil {
			exitGracefully(err) // if this happened we got an unexpected error
		}

		// processing CSV line
		record, err := processLine(headers, line)

		if err != nil {
			fmt.Printf("Line:%sError: %s\n", line, err)
			continue
		}

		// otherwise send the processed record to the writer channel
		writerChannel <- record
	}
}

func exitGracefully(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
}

func check(e error) {
	if e != nil {
		exitGracefully(e)
	}
}

func processLine(headers []string, dataList []string) (map[string]string, error) {
	// validating we are getting the same number of headers and columns
	if len(dataList) != len(headers) {
		return nil, errors.New("line doesnt match headers format. Skipping")
	}

	recordMap := make(map[string]string)

	for i, name := range headers {
		recordMap[name] = dataList[i]
	}
	return recordMap, nil
}
