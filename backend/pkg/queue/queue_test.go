package queue_test

import (
	"strconv"
	"testing"
	"time"

	"pentagi/pkg/queue"
)

func TestQueue_StartStop(t *testing.T) {
	input := make(chan int)
	output := make(chan string)
	workers := 4

	q := queue.NewQueue(input, output, workers, func(i int) (string, error) {
		return strconv.Itoa(i), nil
	})

	if running := q.Running(); running {
		t.Errorf("expected queue to be not running, but it is")
	}

	if err := q.Start(); err != nil {
		t.Errorf("failed to start queue: %v", err)
	}

	if running := q.Running(); !running {
		t.Errorf("expected queue to be running, but it is not")
	}

	if err := q.Stop(); err != nil {
		t.Errorf("failed to stop queue: %v", err)
	}

	if running := q.Running(); running {
		t.Errorf("expected queue to be not running, but it is")
	}
}

func TestQueue_CloseInputChannel(t *testing.T) {
	input := make(chan int)
	output := make(chan string)
	workers := 4

	q := queue.NewQueue(input, output, workers, func(i int) (string, error) {
		return strconv.Itoa(i), nil
	})

	if running := q.Running(); running {
		t.Errorf("expected queue to be not running, but it is")
	}

	if err := q.Start(); err != nil {
		t.Errorf("failed to start queue: %v", err)
	}

	if running := q.Running(); !running {
		t.Errorf("expected queue to be running, but it is not")
	}

	close(input)
	time.Sleep(100 * time.Millisecond)

	if running := q.Running(); running {
		t.Errorf("expected queue to be not running, but it is")
	}
}

func TestQueue_Process(t *testing.T) {
	input := make(chan int)
	output := make(chan string)
	workers := 4

	q := queue.NewQueue(input, output, workers, func(i int) (string, error) {
		return strconv.Itoa(i), nil
	})

	if err := q.Start(); err != nil {
		t.Errorf("failed to start queue: %v", err)
	}

	input <- 42
	result := <-output

	expected := "42"
	if result != expected {
		t.Errorf("unexpected result. expected: %s, got: %s", expected, result)
	}

	if err := q.Stop(); err != nil {
		t.Errorf("failed to stop queue: %v", err)
	}
}

func TestQueue_ProcessOrdering(t *testing.T) {
	input := make(chan int)
	output := make(chan int)
	workers := 4

	q := queue.NewQueue(input, output, workers, func(i int) (int, error) {
		return i + 1, nil
	})

	if err := q.Start(); err != nil {
		t.Errorf("failed to start queue: %v", err)
	}

	go func() {
		for i := 0; i < 100000; i++ {
			input <- i
		}

		if err := q.Stop(); err != nil {
			t.Errorf("failed to stop queue: %v", err)
		}

		close(input)
		close(output)
	}()

	var prev int
	for cur := range output {
		if cur != prev+1 {
			t.Errorf("unexpected result. expected: %d, got: %d", prev+1, cur)
		} else {
			prev = cur
		}
	}
}

func BenchmarkQueue_DefaultWorkers(b *testing.B) {
	simpleBenchmark(b, 0)
}

func BenchmarkQueue_EightWorkers(b *testing.B) {
	simpleBenchmark(b, 8)
}

func BenchmarkQueue_FourWorkers(b *testing.B) {
	simpleBenchmark(b, 4)
}

func BenchmarkQueue_ThreeWorkers(b *testing.B) {
	simpleBenchmark(b, 3)
}

func BenchmarkQueue_TwoWorkers(b *testing.B) {
	simpleBenchmark(b, 2)
}

func BenchmarkQueue_OneWorker(b *testing.B) {
	simpleBenchmark(b, 1)
}

func BenchmarkQueue_OriginalSingleGoroutine(b *testing.B) {
	ch := make(chan struct{})
	input := make(chan int, 100)
	output := make(chan string, 100)
	process := func(i int) (string, error) {
		var res string
		for j := i; j < i+1000; j++ {
			res = strconv.Itoa(i)
		}
		return res, nil
	}

	go func() {
		ch <- struct{}{}
		for i := range input {
			res, _ := process(i)
			output <- res
		}
		close(output)
	}()
	<-ch

	b.ResetTimer()

	go func() {
		for i := 0; i < b.N; i++ {
			input <- i
		}
		close(input)
	}()

	for range output {
	}

	b.StopTimer()
}

func simpleBenchmark(b *testing.B, workers int) {
	input := make(chan int, 100)
	output := make(chan string, 100)
	process := func(i int) (string, error) {
		var res string
		for j := i; j < i+1000; j++ {
			res = strconv.Itoa(i)
		}
		return res, nil
	}
	q := queue.NewQueue(input, output, workers, process)

	if err := q.Start(); err != nil {
		b.Fatalf("failed to start queue: %v", err)
	}

	b.ResetTimer()

	go func() {
		for i := 0; i < b.N; i++ {
			input <- i
		}

		if err := q.Stop(); err != nil {
			b.Errorf("failed to stop queue: %v", err)
		}

		close(input)
		close(output)
	}()

	for range output {
	}

	b.StopTimer()
}
