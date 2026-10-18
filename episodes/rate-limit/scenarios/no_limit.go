package scenarios

import (
	"fmt"
	"sync"
	"time"

	"rate-limit-demo/utils"
)

type NoLimitScenario struct{}

type backendRequest struct {
	client    string
	id        int
	submitted time.Time
	timeout   time.Duration
	result    chan requestResult
}

type requestResult struct {
	waited   time.Duration
	timedOut bool
}

func (NoLimitScenario) Name() string { return "No limit: one noisy client" }
func (NoLimitScenario) Description() string {
	return "A bot fills five backend slots before Alice can run"
}

func (NoLimitScenario) Run() error {
	utils.PrintHeader("Scenario 1: No Limit - One Noisy Client")
	const (
		capacity   = 5
		botCount   = 50
		aliceCount = 3
		work       = 20 * time.Millisecond
		aliceWait  = 150 * time.Millisecond
	)
	fmt.Printf("Backend capacity: %d concurrent requests\n\n", capacity)
	fmt.Printf("Bot:   sends %d requests\nAlice: sends %d requests\n\n", botCount, aliceCount)

	jobs := make(chan backendRequest, botCount+aliceCount)
	slots := make(chan struct{}, capacity)
	started := make(chan struct{}, capacity)
	var workers sync.WaitGroup
	var mu sync.Mutex
	var botStarts int
	botSlots, aliceSlots := 0, 0

	for range capacity {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for r := range jobs {
				waited := time.Since(r.submitted)
				if r.timeout > 0 && waited >= r.timeout {
					r.result <- requestResult{waited: waited, timedOut: true}
					continue
				}
				slots <- struct{}{}
				if r.client == "bot" {
					mu.Lock()
					botStarts++
					firstFive := botStarts <= capacity
					mu.Unlock()
					if firstFive {
						started <- struct{}{}
					}
				}
				time.Sleep(work)
				<-slots
				mu.Lock()
				if r.client == "bot" {
					botSlots++
				} else {
					aliceSlots++
				}
				mu.Unlock()
				r.result <- requestResult{waited: waited}
			}
		}()
	}

	botResults := make(chan requestResult, botCount)
	for i := 1; i <= botCount; i++ {
		jobs <- backendRequest{client: "bot", id: i, submitted: time.Now(), result: botResults}
	}
	// The first five bot requests hold every slot before Alice is queued.
	for range capacity {
		<-started
	}

	aliceResults := make(chan requestResult, aliceCount)
	for i := 1; i <= aliceCount; i++ {
		jobs <- backendRequest{client: "alice", id: i, submitted: time.Now(), timeout: aliceWait, result: aliceResults}
	}
	close(jobs)
	for range botCount {
		<-botResults
	}
	for i := 1; i <= aliceCount; i++ {
		r := <-aliceResults
		if r.timedOut {
			utils.PrintError(fmt.Sprintf("Alice request %d: TIMEOUT after waiting %s", i, r.waited.Round(time.Millisecond)))
		} else {
			utils.PrintWarning(fmt.Sprintf("Alice request %d: waited %s", i, r.waited.Round(time.Millisecond)))
		}
	}
	workers.Wait()

	mu.Lock()
	totalSlots := botSlots + aliceSlots
	share := 0.0
	if totalSlots > 0 {
		share = 100 * float64(botSlots) / float64(totalSlots)
	}
	mu.Unlock()
	fmt.Println()
	utils.PrintWarning(fmt.Sprintf("Bot consumed %d/%d completed backend slots (%.0f%%).", botSlots, totalSlots, share))
	return nil
}
