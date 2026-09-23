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
	fatalIf(json.Unmarshal(normalizeHexEscapes(data), &inputs))
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

// normalizeHexEscapes accepts scanner exports that use \xNN inside JSON
// strings. Standard JSON only permits \u00NN.
func normalizeHexEscapes(data []byte) []byte {
	out := make([]byte, 0, len(data))
	inString := false
	escaped := false
	for i := 0; i < len(data); i++ {
		current := data[i]
		if !inString {
			out = append(out, current)
			if current == '"' {
				inString = true
			}
			continue
		}
		if escaped {
			if current == 'x' && i+2 < len(data) && isHex(data[i+1]) && isHex(data[i+2]) {
				out = append(out, 'u', '0', '0', data[i+1], data[i+2])
				i += 2
			} else {
				out = append(out, current)
			}
			escaped = false
			continue
		}
		out = append(out, current)
		if current == '\\' {
			escaped = true
		} else if current == '"' {
			inString = false
		}
	}
	return out
}

func isHex(value byte) bool {
	return value >= '0' && value <= '9' || value >= 'a' && value <= 'f' || value >= 'A' && value <= 'F'
}

func fatalIf(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
