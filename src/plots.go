package src

import (
	"fmt"
	"image/color"
	"sort"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func PlotCountGroups(path string, countGroups map[int]int, totalTrials int) error {
	var (
		groupA        plotter.Values
		names         []string
		sortedMatches []int
	)
	for numMatches := range countGroups {
		sortedMatches = append(sortedMatches, numMatches)
	}

	sort.Ints(sortedMatches)
	for _, numMatches := range sortedMatches {
		groupA = append(groupA, float64(countGroups[numMatches]))
		if numMatches == 1 {
			names = append(names, fmt.Sprintf("%d match \n(%d)\n ", numMatches, countGroups[numMatches]))
		} else {
			names = append(names, fmt.Sprintf("%d matches \n(%d)\n ", numMatches, countGroups[numMatches]))
		}

	}

	p, err := plot.New()
	if err != nil {
		return err
	}

	p.Title.Text = "Success counts"
	p.Y.Label.Text = fmt.Sprintf("Number of trials (%d total)", totalTrials)

	w := vg.Points(20)

	barsA, err := plotter.NewBarChart(groupA, w)
	if err != nil {
		return err
	}

	barsA.LineStyle.Width = vg.Length(1)
	barsA.Color = color.RGBA{R: 169, G: 169, B: 169}
	barsA.Offset = 0
	p.Add(barsA)
	p.NominalX(names...)

	return p.Save(10*vg.Inch, 6*vg.Inch, path)
}

func PlotCostGroups(path string, costGroups map[float64]int, totalTrials int) error {
	p, err := plot.New()
	if err != nil {
		return err
	}

	points := make(plotter.XYs, len(costGroups))
	var (
		idx         int
		sortedCosts []float64
	)
	for numMatches := range costGroups {
		sortedCosts = append(sortedCosts, numMatches)
	}

	sort.Float64s(sortedCosts)
	for _, group := range sortedCosts {
		points[idx].X = group
		points[idx].Y = float64(costGroups[group])
		idx++
	}
	p.Title.Text = " \n"
	p.X.Label.Text = "Costs"
	p.Y.Label.Text = fmt.Sprintf("Number of trials (%d total)", totalTrials)

	l, err := plotter.NewLine(points)
	if err != nil {
		return err
	}

	l.LineStyle.Width = vg.Points(1)
	l.LineStyle.Color = color.RGBA{B: 255, A: 255}
	p.Add(l)

	return p.Save(10*vg.Inch, 6*vg.Inch, path)
}
