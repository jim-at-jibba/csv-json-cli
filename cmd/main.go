package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
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
