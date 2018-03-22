package internal

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"

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

	var (
		baseScore, matched = list1.Compare(list2, weights)
		baseCount          = len(matched)
		baseResult         = &result{cost: baseScore, matches: matched}
	)
	baseResult.Print()

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
			var (
				shuffled = make([]*Word, len(list2.List))
				perm     = rand.Perm(len(list2.List))
			)
			for i, v := range perm {
				shuffled[v] = list2.List[i]
			}

			jobs <- &job{
				list1:   list1,
				list2:   &SwadeshList{List: shuffled},
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
			if currResult.cost >= baseScore {
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

func worker(id int, jobs chan *job, results chan *result) {
	for args := range jobs {
		cost, matched := args.list1.Compare(args.list2, args.weights)
		results <- &result{cost: cost, matches: matched}
	}
}

type job struct {
	list1   *SwadeshList
	list2   *SwadeshList
	weights Weights
}

type result struct {
	cost    float64
	matches []string
}

func (r *result) Print() {
	var msg string
	for idx, match := range r.matches {
		msg += fmt.Sprintf("Positive pair %d: %s\n", idx, match)
	}
	log.Printf("%s", msg)
	log.Printf("N = %d (number of positive pairs in the original list)\n", len(r.matches))
	log.Printf("S = %f (cost of positive pairs in the original list)\n\n", r.cost)
}
