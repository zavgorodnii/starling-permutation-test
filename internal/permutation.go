package internal

import (
	"log"
	"math/rand"
	"runtime"

	"sync"

	"fmt"

	"github.com/pkg/errors"
)

var (
	scale = runtime.NumCPU()
)

type Summary struct {
	Counts      map[int]int
	Costs       map[float64]int
	TotalCounts int
	TotalCost   int
}

func RunTest(list1, list2 *SwadeshList, weights Weights, trials float64, verbose bool) (summary *Summary, err error) {
	if len(list1.List) != len(list2.List) {
		return nil, errors.Errorf("wordlists have different lengths: %d, %d",
			len(list1.List), len(list2.List))
	}

	baseScore, matched, err := getScore(list1.List, list2.List, weights)
	if err != nil {
		return nil, err
	}

	baseCount := len(matched)
	baseResult := &result{cost: baseScore, matches: matched}
	baseResult.Print("no permutations")

	var (
		jobs    = make(chan *job, scale*2)
		results = make(chan *result, scale*2)
	)
	for workerID := 0; workerID < scale; workerID++ {
		go worker(workerID, jobs, results)
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		for i := 0.; i < trials; i++ {
			shuffled := make([]*Word, len(list1.List))
			perm := rand.Perm(len(list1.List))
			for i, v := range perm {
				shuffled[v] = list1.List[i]
			}

			jobs <- &job{
				list1:   shuffled,
				list2:   list2.List,
				weights: weights,
			}
		}
		wg.Done()
	}(wg)

	wg.Add(1)
	summary = &Summary{
		Counts: map[int]int{},
		Costs:  map[float64]int{},
	}
	go func(wg *sync.WaitGroup, results chan *result) {
		for i := 0.; i < trials; i++ {
			currResult := <-results
			if len(currResult.matches) >= baseCount {
				summary.TotalCounts++
				if verbose {
					currResult.Print()
				}
			}
			if currResult.cost > baseScore {
				summary.TotalCost++
			}
			summary.Counts[len(currResult.matches)]++
			summary.Costs[currResult.cost]++
		}
		wg.Done()
	}(wg, results)

	wg.Wait()

	return summary, nil
}

func getScore(list1, list2 []*Word, weights Weights) (cost float64, matches []string, err error) {
	for idx := 0; idx < len(list1); idx++ {
		word1, word2 := list1[idx], list2[idx]
		if ok, match := word1.Compare(word2); ok {
			cost += weights.GetWeight(word1.SwadeshID)
			matches = append(matches, match)
		}
	}

	return
}

func worker(id int, jobs chan *job, results chan *result) {
	for args := range jobs {
		cost, matched, err := getScore(args.list1, args.list2, args.weights)
		if err != nil {
			log.Fatalf("Worker %d failed: %s", id, err)
		}

		results <- &result{cost: cost, matches: matched}
	}
}

type job struct {
	list1   []*Word
	list2   []*Word
	weights Weights
}

type result struct {
	cost    float64
	matches []string
}

func (r *result) Print(additionalInfo ...string) {
	var msg string
	for idx, match := range r.matches {
		msg += fmt.Sprintf("Positive pair %d: %s\n", idx, match)
	}
	log.Printf("%s", msg)
	log.Printf("N = %d (number of positive pairs in the original list)\n", len(r.matches))
	log.Printf("S = %f (cost of positive pairs in the original list)\n\n", r.cost)
}
