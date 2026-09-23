package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Cunyule/AI-coding/internal/model"
)

func main() {
	inputPath := flag.String("input", "testdata/input.json", "input JSON file")
	serverURL := flag.String("server", "http://127.0.0.1:8080", "fingerprint server base URL")
	timeout := flag.Duration("timeout", 15*time.Second, "request timeout")
	compact := flag.Bool("compact", false, "print compact JSON")
	flag.Parse()

	data, err := os.ReadFile(*inputPath)
	fatalIf(err)
	var inputs []model.ScanRecord
	fatalIf(json.Unmarshal(data, &inputs))
	payload, err := json.Marshal(inputs)
	fatalIf(err)

	request, err := http.NewRequest(http.MethodPost, *serverURL+"/fingerprint", bytes.NewReader(payload))
	fatalIf(err)
	request.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Timeout: *timeout}).Do(request)
	fatalIf(err)
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 16<<20))
	fatalIf(err)
	if response.StatusCode != http.StatusOK {
		fatalIf(fmt.Errorf("server returned %s: %s", response.Status, string(body)))
	}
	var results []model.FingerprintResult
	fatalIf(json.Unmarshal(body, &results))
	if *compact {
		body, err = json.Marshal(results)
	} else {
		body, err = json.MarshalIndent(results, "", "  ")
	}
	fatalIf(err)
	fmt.Println(string(body))
}

func fatalIf(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
