package internal

import (
	"log"
	"sort"

	"github.com/pkg/errors"
	"github.com/tealeg/xlsx"
)

type Weights interface {
	GetWeight(swadeshID int) float64
}

type WeightsStore struct {
	swadeshIDToWeight map[int]float64
}

func NewWeightsStore(weightsPath string, verbose bool) (Weights, error) {
	classesFile, err := xlsx.OpenFile(weightsPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s", weightsPath)
	}

	if len(classesFile.Sheets) != 1 {
		return nil, errors.New("single sheet is expected")
	}

	var classIDtoWeight = map[int]float64{}
	for idx := 1; idx < len(classesFile.Sheets[0].Rows); idx++ {
		row := classesFile.Sheets[0].Rows[idx]
		if len(row.Cells) < 3 {
			return nil, errors.Errorf("row %d has less than 2 cells", idx)
		}

		swadeshID, err := row.Cells[0].Int()
		if err != nil {
			return nil, errors.Wrapf(err, "swadesh ID from row %d is not a float value", idx)
		}

		classWeight, err := row.Cells[2].Float()
		if err != nil {
			return nil, errors.Wrapf(err, "weight from row %d is not a float value", idx)
		}

		classIDtoWeight[swadeshID] = classWeight
	}

	if verbose {
		var sortedIDs []int
		for classID := range classIDtoWeight {
			sortedIDs = append(sortedIDs, classID)
		}
		sort.Ints(sortedIDs)
		log.Println("Found swadesh ID weights:")
		for _, id := range sortedIDs {
			log.Printf("%d\t<-->\t%f\n", id, classIDtoWeight[id])
		}
	}

	return &WeightsStore{swadeshIDToWeight: classIDtoWeight}, nil
}

func (w *WeightsStore) GetWeight(swadeshID int) float64 {
	if weight, ok := w.swadeshIDToWeight[swadeshID]; ok {
		return weight
	}

	return 1.0
}

type DefaultWeightsStore struct{}

func (w *DefaultWeightsStore) GetWeight(swadeshID int) float64 {
	return 1.0
}
