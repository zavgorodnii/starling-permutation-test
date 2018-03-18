package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/starling-permutation-test/internal"
)

var (
	soundsPath = flag.String(
		"sounds", "./data/sounds.xlsx", "path to file containing sound classes")
	wordlistsPath = flag.String(
		"wordlists", "./data/wordlists.xlsx", "path to file containing wordlists")
	wordlistsPathA = flag.String(
		"set_a", "", "path to file containing wordlists for A (triggers AB mode)")
	wordlistsPathB = flag.String(
		"set_b", "", "path to file containing wordlists for B (triggers AB mode)")
	weightsPath = flag.String(
		"weights", "", "path to file containing class weights")
	numTrials  = flag.Int("num_trials", 1000000, "number of trials")
	verbose    = flag.Bool("verbose", false, "verbose output")
	outputPath = flag.String("output", "", "path to output file (stdout if not specified)")
	plotPath   = flag.String("count_groups_plot", "",
		"path to file with count groups plot")
	weightedPlotPath = flag.String("cost_groups_plot", "",
		"path to file with cost groups plot")
	consonantPath = flag.String("consonant", "", "path to file with consonant encodings")
	abMode        bool
)

func init() {
	flag.Parse()

	if len(*wordlistsPathA) > 0 || len(*wordlistsPathB) > 0 {
		if len(*wordlistsPathA) == 0 || len(*wordlistsPathB) == 0 {
			log.Println("Both `--wordlist_a` and `--wordlist_b` must be specified, exiting")
			os.Exit(1)
		}

		abMode = true
	}

	rand.Seed(time.Now().UnixNano())
}

func main() {
	if len(*outputPath) > 0 {
		w, err := os.OpenFile(*outputPath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Printf("Failed to open %s for writing (using stdout): %s", *outputPath, err)
		} else {
			log.SetOutput(w)
		}
		defer w.Close()
	}
	log.SetFlags(0)

	var weights internal.Weights = &internal.DefaultWeightsStore{}
	if len(*weightsPath) > 0 {
		if weightsStore, err := internal.NewWeightsStore(*weightsPath, *verbose); err != nil {
			log.Printf("Failed to open weight file %s (%s), using defaults weights (1.0)", *weightsPath, err)
		} else {
			weights = weightsStore
		}
	}

	if abMode {
		runPermutationTestAB(weights)
	} else {
		runPermutationTest(weights)
	}
}

func runPermutationTest(weights internal.Weights) {
	decoder, err := internal.NewSoundClassesInfo(*soundsPath)
	if err != nil {
		log.Println("Failed to load sound classes info:", err)
		return
	}

	wordlists, err := decoder.Decode(*wordlistsPath)
	if err != nil {
		log.Println("Failed to decode wordlists:", err)
		return
	}

	runTest(wordlists[0], wordlists[1], weights)
}

func runPermutationTestAB(weights internal.Weights) {
	decoder, err := internal.NewSoundClassesInfo(*soundsPath)
	if err != nil {
		log.Println("Failed to load sound classes info:", err)
		return
	}

	wordlistsA, err := decoder.Decode(*wordlistsPathA)
	if err != nil {
		log.Println("Failed to decode wordlists A:", err)
		return
	}

	var combinedA = wordlistsA[0]
	for idx := 1; idx < len(wordlistsA); idx++ {
		combinedA = combinedA.Combine(wordlistsA[idx])
	}

	wordlistsB, err := decoder.Decode(*wordlistsPathB)
	if err != nil {
		log.Println("Failed to decode wordlists B:", err)
		return
	}

	var combinedB = wordlistsB[0]
	for idx := 1; idx < len(wordlistsB); idx++ {
		combinedB = combinedB.Combine(wordlistsB[idx])
	}

	runTest(combinedA, combinedB, weights)
}

func runTest(l1, l2 *internal.SwadeshList, weights internal.Weights) {
	summary, err := internal.RunTest(l1, l2, weights, float64(*numTrials), *verbose)
	if err != nil {
		log.Println("Failed to run permutation test:", err)
		return
	}

	var sortedCountGroups []int
	for numMatches := range summary.Counts {
		sortedCountGroups = append(sortedCountGroups, numMatches)
	}
	sort.Ints(sortedCountGroups)
	for _, countGroup := range sortedCountGroups {
		log.Printf("k = %d:\t%d trial(s)\n", countGroup, summary.Counts[countGroup])
	}
	log.Printf("P (counts) = %d / %d = %f\n\n", summary.TotalCounts, *numTrials,
		float64(summary.TotalCounts)/float64(*numTrials))

	if len(*weightsPath) > 0 {
		var sortedCosts []float64
		for numMatches := range summary.Costs {
			sortedCosts = append(sortedCosts, numMatches)
		}
		sort.Float64s(sortedCosts)
		for _, costGroup := range sortedCosts {
			log.Printf("s = %.3f: %d trial(s)\n", costGroup, summary.Costs[costGroup])
		}
		log.Printf("P (costs) = %d / %d = %f\n", summary.TotalCost, *numTrials,
			float64(summary.TotalCost)/float64(*numTrials))

		if len(*weightedPlotPath) > 0 {
			if err := internal.PlotCostGroups(*weightedPlotPath, summary.Costs, *numTrials); err != nil {
				log.Printf("Failed to plot cost groups: %s", err)
			} else {
				log.Printf("Cost groups plot saved at %s", *weightedPlotPath)
			}
		}
	}

	if len(*plotPath) > 0 {
		if err := internal.PlotCountGroups(*plotPath, summary.Counts, *numTrials); err != nil {
			log.Printf("Failed to plot count groups: %s", err)
		} else {
			log.Printf("Count groups plot saved at %s", *plotPath)
		}
	}

	if len(*consonantPath) > 0 {
		consonantW, err := os.OpenFile(*consonantPath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Printf("Failed to open %s for writing (using stdout): %s", *outputPath, err)
		} else {
			log.SetOutput(consonantW)
			l1.PrintTransformations()
			l2.PrintTransformations()
		}
	}
}
