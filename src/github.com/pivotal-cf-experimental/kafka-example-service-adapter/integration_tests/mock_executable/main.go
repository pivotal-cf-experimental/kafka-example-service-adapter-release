package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	data, err := json.Marshal(os.Args)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(os.Getenv("TEST_PARAMS_FILE_NAME"), data, os.ModePerm); err != nil {
		panic(err)
	}

	if stderr := os.Getenv("TEST_EXECUTABLE_SHOULD_FAIL"); stderr != "" {
		fmt.Fprintln(os.Stderr, stderr)
		os.Exit(1)
	}
}
