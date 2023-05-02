package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

func getIterations() int {
	if value := os.Getenv("ITERATIONS"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil {
			return parsed
		}
	}
	return 100
}

func main() {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to connect to database: %+v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	ctx := context.Background()
	iterations := getIterations()

	start := time.Now()
	var bitbucket int
	var queryStart time.Time
	durations := make([]time.Duration, iterations)
	for x := 0; x < iterations; x++ {
		queryStart = time.Now()
		_ = conn.QueryRow(ctx, "select 1").Scan(&bitbucket)
		durations[x] = time.Since(queryStart)
	}

	totalElapsed := time.Since(start)

	fmt.Printf("iterations=%d\n", iterations)
	fmt.Printf("elapsedTotal=%v\n", totalElapsed.Round(time.Millisecond))
	fmt.Printf("queryElapsedP95=%v\n", Percentile(durations, 95.0).Round(time.Millisecond))
	fmt.Printf("queryElapsedP50=%v\n", Percentile(durations, 50.0).Round(time.Millisecond))
}

// Percentile finds the relative standing in a slice of floats.
// `percent` should be given on the interval [0,100.0).
func Percentile[T Operatable](input []T, percent float64) (output T) {
	if len(input) == 0 {
		return
	}
	output = PercentileSorted(CopySort(input), percent)
	return
}

// PercentileSorted finds the relative standing in a sorted slice of floats.
// `percent` should be given on the interval [0,100.0).
func PercentileSorted[T Operatable](sortedInput []T, percent float64) (percentile T) {
	if len(sortedInput) == 0 {
		return
	}
	index := (percent / 100.0) * float64(len(sortedInput))
	i := int(math.RoundToEven(index))
	if index == float64(int64(index)) {
		percentile = (sortedInput[i-1] + sortedInput[i]) / 2.0
	} else {
		percentile = sortedInput[i-1]
	}
	return percentile
}

// CopySort copies and sorts a slice ascending.
func CopySort[T Ordered](input []T) (output []T) {
	output = make([]T, len(input))
	copy(output, input)
	sort.Slice(output, func(i, j int) bool {
		return output[i] < output[j]
	})
	return
}

// Ordered are types that can be sorted.
type Ordered interface {
	string | ~int | ~float64 | time.Duration
}

// Operatable are types that can be mathed.
type Operatable interface {
	~int | ~float64 | time.Duration
}
