package main

import (
	"fmt"
	"sync"
	"time"
)

// Model
type model struct {
	count          int        // Counter
	bufferSize     int        // Buffer size
	mtx            sync.Mutex // Mutex lock
	consumerWakeup *sync.Cond // Consumer wake-up condition
	producerWakeup *sync.Cond // Producer wake-up condition
}

// Producer receiver
type producer struct{}

// Consumer receiver
type consumer struct{}

// Progress bar
func printProgress(count, bufferSize int) {
	fmt.Printf("\rProducer: [")
	for i := 0; i < count; i++ {
		fmt.Print("█") // Producer's progress
	}
	for i := count; i < bufferSize; i++ {
		fmt.Print("░") // Available production space
	}

	fmt.Print("] (", count, "/", bufferSize, ")   Consumer: [")
	for i := 0; i < bufferSize-count; i++ {
		fmt.Print("█") // Consumer's progress
	}
	for i := bufferSize - count; i < bufferSize; i++ {
		fmt.Print("░") // Available consumption space
	}
	fmt.Print("] (", bufferSize-count, "/", bufferSize, ")")
}

// Producer produces data
func (p producer) produce(c *model) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// Wait until there is space in the buffer
	for c.count == c.bufferSize {
		fmt.Println("Buffer full, producer waiting...")
		c.producerWakeup.Wait()
	}

	// Produce data
	c.count++
	printProgress(c.count, c.bufferSize)

	// Wake up consumer
	c.consumerWakeup.Signal()
}

// Consumer consumes data
func (c consumer) consume(m *model) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	// Wait until there is data in the buffer
	for m.count == 0 {
		fmt.Println("Buffer empty, consumer waiting...")
		m.consumerWakeup.Wait()
	}

	// Consume data
	m.count--
	printProgress(m.count, m.bufferSize)

	// Wake up producer
	m.producerWakeup.Signal()
}

func main() {
	// Initialize model
	c := &model{
		count:      0, // Initial counter value
		bufferSize: 5, // Set buffer size
	}
	// Initialize consumer wake-up condition
	c.consumerWakeup = sync.NewCond(&c.mtx)
	// Initialize producer wake-up condition
	c.producerWakeup = sync.NewCond(&c.mtx)

	// Number of producers
	producerCount := 1
	// Number of consumers
	consumerCount := 1

	var wg sync.WaitGroup

	// Start producers
	for i := 0; i < producerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				time.Sleep(time.Millisecond * 100)
				producer{}.produce(c)
			}
		}(i)
	}

	// Start consumers
	for i := 0; i < consumerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				time.Sleep(time.Millisecond * 150)
				consumer{}.consume(c)
			}
		}(i)
	}

	// Wait for all producers and consumers to complete
	wg.Wait()
	fmt.Printf("\ncurrent count: %v\n", c.count)
	fmt.Println("All producers and consumers completed.")
}
