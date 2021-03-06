package main

import (
	"flag"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func Test_getFileData(t *testing.T) {
	tests := []struct {
		name    string
		want    inputFile
		wantErr bool
		osArgs  []string
	}{
		{"Default parameters", inputFile{"test.csv", "comma", false}, false, []string{"cmd", "test.csv"}},
		{"No parameters", inputFile{}, true, []string{"cmd"}},
		{"Semicolon enabled", inputFile{"test.csv", "semicolon", false}, false, []string{"cmd", "--separator=semicolon", "test.csv"}},
		{"Pretty enabled", inputFile{"test.csv", "comma", true}, false, []string{"cmd", "--pretty", "test.csv"}},
		{"Pretty and semicolon enabled", inputFile{"test.csv", "semicolon", true}, false, []string{"cmd", "--pretty", "--separator=semicolon", "test.csv"}},
		{"Separator not identified", inputFile{}, true, []string{"cmd", "--separator=pipe", "test.csv"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save the original os.Args
			actualOsArgs := os.Args
			defer func() {
				os.Args = actualOsArgs                                           // restoring the original
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // Resetting flags
			}()

			os.Args = tt.osArgs       // setting specific command args for this test
			got, err := getFileData() // running function we want to test
			if (err != nil) != tt.wantErr {
				t.Errorf("getFileData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFileData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkIfValidFile(t *testing.T) {
	// create tmp and empty csv
	tmpFile, err := ioutil.TempFile("", "text*.csv")

	if err != nil {
		panic(err)
	}

	// Once all the tests are done, delete the temp file
	defer os.Remove(tmpFile.Name())

	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"File does exist", args{filename: tmpFile.Name()}, true, false},
		{"File does not exist", args{filename: "nowehere/test.csv"}, false, true},
		{"File is not csv", args{filename: "test.txt"}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkIfValidFile(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkIfValidFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checkIfValidFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_processCsvFile(t *testing.T) {
	type args struct {
		fileData      inputFile
		writerChannel chan<- map[string]string
	}
	wantMapSlice := []map[string]string{
		{"COL1": "1", "COL2": "2", "COL3": "3"},
		{"COL1": "4", "COL2": "5", "COL3": "6"},
	}

	tests := []struct {
		name      string // The name of the test
		csvString string // The content of our tested CSV file
		separator string // The separator used for each test case
	}{
		{"Comma separator", "COL1,COL2,COL3\n1,2,3\n4,5,6\n", "comma"},
		{"Semicolon separator", "COL1;COL2;COL3\n1;2;3\n4;5;6\n", "semicolon"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Creating a CSV temp file for testing
			tmpfile, err := ioutil.TempFile("", "test*.csv")
			check(err)

			defer os.Remove(tmpfile.Name())            // Removing the CSV test file before living
			_, err = tmpfile.WriteString(tt.csvString) // Writing the content of the CSV test file
			tmpfile.Sync()                             // Persisting data on disk
			// Defining the inputFile struct that we're going to use as one parameter of our function
			testFileData := inputFile{
				filepath:  tmpfile.Name(),
				pretty:    false,
				separator: tt.separator,
			}
			// Defining the writerChanel
			writerChannel := make(chan map[string]string)
			// Calling the targeted function as a go routine
			go processCsvFile(testFileData, writerChannel)
			// Iterating over the slice containing the expected map values
			for _, wantMap := range wantMapSlice {
				record := <-writerChannel                // Waiting for the record that we want to compare
				if !reflect.DeepEqual(record, wantMap) { // Making the corresponding test assertion
					t.Errorf("processCsvFile() = %v, want %v", record, wantMap)
				}
			}
		})
	}
}
