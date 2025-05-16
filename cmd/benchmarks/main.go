package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

const MAX_KEEP = 30

// Root structure containing all the data
type Benchmark struct {
	Data []DataPoint `json:"data"`
}

// DataPoint represents each measurement point with its label and metrics
type DataPoint struct {
	XLabel    string        `json:"x-label"`
	Timestamp string        `json:"timestamp"`
	Points    []MetricPoint `json:"points"`
}

// MetricPoint represents an individual measurement
type MetricPoint struct {
	Name  string  `json:"name"`
	Unit  string  `json:"unit"`
	Value float64 `json:"value"`
	Label string  `json:"label"`
}

type BenchmarkResult struct {
	Name  string  `json:"name"`
	Unit  string  `json:"unit"`
	Value float64 `json:"value"`
	Extra string  `json:"extra"`
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: benchmarks <data.json> <output.json>")
	}

	dataJson := os.Args[1]
	outputJson := os.Args[2]

	rotelRelease := os.Getenv("ROTEL_RELEASE")
	if rotelRelease == "" {
		log.Fatalf("Must set ROTEL_RELEASE")
	}

	otelSha := os.Getenv("OTEL_SHA")
	if otelSha == "" {
		log.Fatalf("Must set OTEL_SHA")
	}

	benchmarkData := Benchmark{
		Data: nil,
	}
	f, err := os.Open(dataJson)
	if err == nil {
		// The first time this may not exist, that's fine

		// Read the file content
		buf, err := io.ReadAll(f)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}

		// Parse the JSON into our struct
		if err := json.Unmarshal(buf, &benchmarkData); err != nil {
			log.Fatalf("Failed to parse benchmark JSON: %v", err)
		}
		_ = f.Close()
	}

	f, err = os.Open(outputJson)
	if err != nil {
		log.Fatalf("Unable to open %s: %v", outputJson, err)
	}

	buf, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var results []BenchmarkResult
	// Parse the JSON into our struct
	if err := json.Unmarshal(buf, &results); err != nil {
		log.Fatalf("Failed to parse results JSON: %v", err)
	}
	_ = f.Close()

	dp := DataPoint{
		XLabel:    fmt.Sprintf("%s - %s", otelSha, rotelRelease),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Points:    nil,
	}

	for _, result := range results {
		sp := strings.Split(result.Extra, "/")
		if len(sp) != 2 {
			log.Fatalf("invalid extra benchmark result data: %s", result.Extra)
		}

		labelSp := strings.Split(sp[1], " - ")
		if len(labelSp) != 2 {
			log.Fatalf("invalid label: %s", sp[1])
		}

		dp.Points = append(dp.Points, MetricPoint{
			Name:  strings.TrimPrefix(fmt.Sprintf("%s - %s", sp[0], result.Name), "Rotel"),
			Unit:  result.Unit,
			Value: result.Value,
			Label: labelSp[0],
		})
	}

	benchmarkData.Data = append(benchmarkData.Data, dp)
	if len(benchmarkData.Data) > MAX_KEEP {
		benchmarkData.Data = benchmarkData.Data[len(benchmarkData.Data)-MAX_KEEP:]
	}

	jsonOut, err := json.Marshal(&benchmarkData)
	if err != nil {
		log.Fatalf("failed to marshal test data: %v", err)
	}

	f, err = os.OpenFile(dataJson, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("failed to open data file for writing: %v", err)
	}

	_, err = f.Write(jsonOut)
	if err != nil {
		log.Fatalf("failed to write output file: %v", err)
	}

	if f.Close() != nil {
		log.Fatalf("failed to close output file: %v", err)
	}
}
