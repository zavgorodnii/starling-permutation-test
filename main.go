package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"

	"io/ioutil"
	"runtime/debug"

	"github.com/starling-permutation-test/internal"
)

const (
	stackTracePath = "stack.trace"
)

var (
	soundsPath = flag.String(
		"sounds", "./data/sounds.xlsx", "path to file containing sound classes")
	wordlistsPath = flag.String(
		"wordlists", "./data/wordlists.xlsx", "path to file containing wordlists")
	setA             = flag.String("set_a", "", "path to file containing wordlists for A (triggers AB mode)")
	setB             = flag.String("set_b", "", "path to file containing wordlists for B (triggers AB mode)")
	weightsPath      = flag.String("weights", "", "path to file containing class weights")
	numTrials        = flag.Int("num_trials", 1000000, "number of trials")
	verbose          = flag.Bool("verbose", false, "verbose output")
	outputPath       = flag.String("output", "", "path to output file (stdout if not specified)")
	plotPath         = flag.String("count_groups_plot", "", "path to file with count groups plot")
	weightedPlotPath = flag.String("cost_groups_plot", "", "path to file with cost groups plot")
	consonantPath    = flag.String("consonant", "", "path to file with consonant encodings")
	lang1            = flag.String("lang_1", "", "first language to compare (optional)")
	lang2            = flag.String("lang_2", "", "second language to compare (optional)")
	allPairs         = flag.Bool("all_pairs", false, "compare all pairs for two wordlists")
	abMode           bool
)

func init() {
	flag.Parse()

	if len(*setA) > 0 || len(*setB) > 0 {
		if len(*setA) == 0 || len(*setB) == 0 {
			log.Println("Both `--wordlist_a` and `--wordlist_b` must be specified, exiting")
			os.Exit(1)
		}

		abMode = true
	}

	rand.Seed(time.Now().UnixNano())
}

func main() {
	defer func() {
		if e := recover(); e != nil {
			currDir, _ := os.Getwd()
			log.Printf("Program crashed, see %s/%s for details\n", currDir, stackTracePath)
			ioutil.WriteFile(stackTracePath, []byte(debug.Stack()), 0666)
		}
	}()

	if len(*outputPath) > 0 {
		os.Remove(*outputPath)
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

	var selectedLangs map[string]bool
	if len(*lang1) > 0 && len(*lang2) > 0 {
		selectedLangs = map[string]bool{*lang1: true, *lang2: true}
	}
	wordlists, err := decoder.Decode(*wordlistsPath, selectedLangs)
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

	wordlistsA, err := decoder.Decode(*setA, nil)
	if err != nil {
		log.Println("Failed to decode wordlists A:", err)
		return
	}

	var combinedA = wordlistsA[0]
	for idx := 1; idx < len(wordlistsA); idx++ {
		combinedA = combinedA.Combine(wordlistsA[idx])
	}

	wordlistsB, err := decoder.Decode(*setB, nil)
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
	summary, err := internal.CompareWordlists(l1, l2, weights, float64(*numTrials), *allPairs, *verbose)
	if err != nil {
		log.Println("Failed to run permutation test:", err)
		return
	}

	if *allPairs {
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
		os.Remove(*consonantPath)
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
