package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/starling-permutation-test/src"
)

const (
	stackTracePath = "stack.trace"
)

var (
	soundsPath       = flag.String("sounds", "./data/sounds.xlsx", "path to file containing sound classes")
	wordlistsPath    = flag.String("wordlists", "./data/wordlists.xlsx", "path to file containing wordlists")
	setA             = flag.String("set_a", "", "path to file containing wordlists for A (triggers AB mode)")
	setB             = flag.String("set_b", "", "path to file containing wordlists for B (triggers AB mode)")
	weightsPath      = flag.String("weights", "", "path to file containing class weights")
	outputPath       = flag.String("output", "", "path to output file (stdout if not specified)")
	plotPath         = flag.String("count_groups_plot", "", "path to file with count groups plot")
	weightedPlotPath = flag.String("cost_groups_plot", "", "path to file with cost groups plot")
	consonantPath    = flag.String("consonants", "", "path to file with consonant encodings")
	lang1            = flag.String("lang_1", "", "first language to compare (optional)")
	lang2            = flag.String("lang_2", "", "second language to compare (optional)")
	allPairs         = flag.Bool("all_pairs", false, "compare each wordlist in file")
	verbose          = flag.Bool("verbose", false, "verbose output")
	numTrials        = flag.Int("num_trials", 1000000, "number of trials")
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
	log.SetFlags(0)
}

func main() {
	defer func() {
		if e := recover(); e != nil {
			currDir, _ := os.Getwd()
			fmt.Printf("Program crashed, see %s/%s for details\n", currDir, stackTracePath)
			_ = ioutil.WriteFile(stackTracePath, []byte(debug.Stack()), 0666)
		}
	}()

	var weights src.Weights = &src.DefaultWeightsStore{}
	if len(*weightsPath) > 0 {
		if weightsStore, err := src.NewWeightsStore(*weightsPath, *verbose); err != nil {
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

func runPermutationTest(weights src.Weights) {
	decoder, err := src.NewSoundClassesDecoder(*soundsPath)
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

	if *allPairs {
		for i := 0; i < len(wordlists); i++ {
			printConsonants(wordlists[i])
			for j := i; j < len(wordlists); j++ {
				if i != j {
					wFile := setupOutput(wordlists[i], wordlists[j])
					if len(*weightsPath) > 0 {
						runTestWeighted(wordlists[i], wordlists[j], weights)
					} else {
						runTest(wordlists[i], wordlists[j], weights)
					}
					if wFile != nil {
						wFile.Close()
					}
				}
			}
		}
	} else {
		wFile := setupOutput(wordlists[0], wordlists[1])
		if len(*weightsPath) > 0 {
			runTestWeighted(wordlists[0], wordlists[1], weights)
		} else {
			runTest(wordlists[0], wordlists[1], weights)
		}
		if wFile != nil {
			wFile.Close()
		}
		printConsonants(wordlists[0])
		printConsonants(wordlists[1])
	}
}

func setupOutput(l1, l2 *src.Wordlist) *os.File {
	if len(*outputPath) > 0 {
		var expOutputPath = expandPath(*outputPath, l1, l2)
		os.Remove(expOutputPath)
		w, err := os.OpenFile(expOutputPath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Printf("Failed to open %s for writing (using stdout): %s", *outputPath, err)
		} else {
			log.SetOutput(w)
		}

		return w
	}

	return nil
}

func printConsonants(l1 *src.Wordlist) {
	if len(*consonantPath) > 0 {
		var expConsonantPath = expandPath(*consonantPath, l1, &src.Wordlist{})
		os.Remove(expConsonantPath)
		consonantW, err := os.OpenFile(expConsonantPath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Printf("Failed to open %s for writing (using stdout): %s", *outputPath, err)
			log.SetOutput(os.Stdout)
		} else {
			log.SetOutput(consonantW)
		}

		l1.PrintTransformations()
	}
}

func runPermutationTestAB(weights src.Weights) {
	decoder, err := src.NewSoundClassesDecoder(*soundsPath)
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

	wFile := setupOutput(combinedA, combinedB)
	if len(*weightsPath) > 0 {
		runTestWeighted(combinedA, combinedB, weights)
	} else {
		runTest(combinedA, combinedB, weights)
	}
	if wFile != nil {
		wFile.Close()
	}
	printConsonants(combinedA)
	printConsonants(combinedB)
}

func runTestWeighted(l1, l2 *src.Wordlist, weights src.Weights) {
	maxCost, group1, group2 := runTest(l1, l2, weights), l1.Group, l2.Group
	if cost := runTest(l2, l1, weights); cost > maxCost {
		maxCost, group1, group2 = cost, l2.Group, l1.Group
	}

	log.Printf("\n[FINAL] Max P(costs) = %f (%s, %s)", maxCost, group1, group2)
}

func runTest(l1, l2 *src.Wordlist, weights src.Weights) (weightedCost float64) {
	log.Printf("\n[Comparing %s with %s]", l1.Group, l2.Group)

	summary, err := src.CompareWordlists(l1, l2, weights, float64(*numTrials), *verbose)
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

		weightedCost = float64(summary.TotalCost) / float64(*numTrials)
		log.Printf("P (costs) = %d / %d = %f\n", summary.TotalCost, *numTrials, weightedCost)

		if len(*weightedPlotPath) > 0 {
			var expWeightedPlotPath = expandPlotPath(*weightedPlotPath, l1, l2)
			os.Remove(expWeightedPlotPath)
			if err := src.PlotCostGroups(expWeightedPlotPath, summary.Costs, *numTrials); err != nil {
				log.Printf("Failed to plot cost groups: %s", err)
			} else {
				log.Printf("Cost groups plot saved at %s", *weightedPlotPath)
			}
		}
	}

	if len(*plotPath) > 0 {
		var expPlotPath = expandPlotPath(*plotPath, l1, l2)
		os.Remove(expPlotPath)
		if err := src.PlotCountGroups(expPlotPath, summary.Counts, *numTrials); err != nil {
			log.Printf("Failed to plot count groups: %s", err)
		} else {
			log.Printf("Count groups plot saved at %s", *plotPath)
		}
	}

	return weightedCost
}

func expandPath(path string, l1, l2 *src.Wordlist) string {
	return strings.Split(path, ".txt")[0] + fmt.Sprintf("_%s_%s", l1.Group, l2.Group) + ".txt"
}

func expandPlotPath(path string, l1, l2 *src.Wordlist) string {
	var extension string
	if strings.Contains(path, ".svg") {
		extension = ".svg"
	} else if strings.Contains(path, ".png") {
		extension = ".png"
	} else if strings.Contains(path, ".jpeg") {
		extension = ".jpeg"
	} else {
		log.Panicf("Plot path should contain extension (one of `.svg`, `.png`, `.jpeg`")
	}
	return strings.Split(path, extension)[0] + fmt.Sprintf("_%s_%s", l1.Group, l2.Group) + extension
}
