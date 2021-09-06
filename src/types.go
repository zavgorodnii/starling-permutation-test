package src

import (
	"fmt"
	"log"
	"strings"
)

type Wordlist struct {
	Group string
	List  []*Word
}

func (l *Wordlist) Compare(other *Wordlist, weights Weights) (cost float64, matches []string) {
	for idx := range l.List {
		word1, word2 := l.List[idx], other.List[idx]
		if ok, match := word1.Compare(word2); ok {
			cost += weights.GetWeight(word1.SwadeshID)
			matches = append(matches, match)
		}
	}

	return
}

func (l *Wordlist) Combine(other *Wordlist) *Wordlist {
	var (
		merged []*Word
		l1, l2 = l.List, other.List
		out    = &Wordlist{Group: fmt.Sprintf("%s, %s", l.Group, other.Group)}
	)
	for len(l1) > 0 && len(l2) > 0 {
		if l1[0].SwadeshID < l2[0].SwadeshID {
			merged = append(merged, l1[0].DeepCopy())
			l1 = l1[1:]
		} else if l1[0].SwadeshID > l2[0].SwadeshID {
			merged = append(merged, l2[0].DeepCopy())
			l2 = l2[1:]
		} else {
			w1, w2 := l1[0].DeepCopy(), l2[0].DeepCopy()
			w1.Group = fmt.Sprintf("%s, %s", w1.Group, w2.Group)
			w1.Forms = append(w1.Forms, w2.Forms...)
			w1.CleanForms = append(w1.CleanForms, w2.CleanForms...)
			w1.DecodedForms = append(w1.DecodedForms, w2.DecodedForms...)
			merged = append(merged, w1)
			l1 = l1[1:]
			l2 = l2[1:]
		}
	}

	if len(l1) > 0 {
		merged = append(merged, l1...)
	}

	if len(l2) > 0 {
		merged = append(merged, l2...)
	}

	out.List = merged

	return out
}

func (l *Wordlist) PrintTransformations() {
	for _, word := range l.List {
		word.PrintTransformations()
	}
}

type Word struct {
	Group        string
	SwadeshID    int
	SwadeshWord  string
	BroomedSymbols []string
	Forms        []string
	CleanForms   []string
	DecodedForms []string
}

func (w *Word) PrintTransformations() {
	var (
		forms  = strings.Join(w.Forms, "; ")
		parsed []string
	)

	for idx, cleanForm := range w.CleanForms {
		parsed = append(parsed, fmt.Sprintf("%s\t-->\t%s (Total %d symbols)",
			cleanForm, w.DecodedForms[idx], len(w.DecodedForms[idx])))
	}
	formatted := fmt.Sprintf("")
	if len(w.BroomedSymbols)>0{
		formatted = fmt.Sprintf("%s (%s)\nSeen as: %s\n[Transformed]\n%s\n[Broomed symbols] %s\n",
			w.SwadeshWord, w.Group, forms, strings.Join(parsed, "\n"), w.BroomedSymbols)
	} else {
		formatted = fmt.Sprintf("%s (%s)\nSeen as: %s\n[Transformed]\n%s\n",
			w.SwadeshWord, w.Group, forms, strings.Join(parsed, "\n"))
	}

	log.Println(formatted)
}

func (w *Word) Compare(other *Word) (bool, string) {
	for idx1, form1 := range w.DecodedForms {
		for idx2, form2 := range other.DecodedForms {
			if len(form1) == 0 || len(form2) == 0 {
				return false, ""
			}

			var (
				isEqual  = true
				upperIdx = 2
			)

			if len(form1) != len(form2) && (len(form1) < upperIdx || len(form2) < upperIdx) {
				isEqual = false
			} else {
				for i := 0; i < upperIdx; i++ {
					if form1[i] != form2[i] {
						isEqual = false
						break
					}
				}
			}

			if isEqual {
				return true, fmt.Sprintf("%d %s: %s - %s", w.SwadeshID, w.SwadeshWord,
					w.CleanForms[idx1], other.CleanForms[idx2])
			}
		}
	}

	return false, ""
}

func (w *Word) DeepCopy() *Word {
	formsCopy := make([]string, len(w.Forms))
	copy(formsCopy, w.Forms)

	cleanFormsCopy := make([]string, len(w.CleanForms))
	copy(cleanFormsCopy, w.CleanForms)

	decodedFormsCopy := make([]string, len(w.DecodedForms))
	copy(decodedFormsCopy, w.DecodedForms)

	BroomedSymbolsCopy := make([]string, len(w.BroomedSymbols))
	copy(formsCopy, w.BroomedSymbols)

	return &Word{
		Group:        w.Group,
		SwadeshID:    w.SwadeshID,
		SwadeshWord:  w.SwadeshWord,
		BroomedSymbols: 			BroomedSymbolsCopy,
		Forms:        formsCopy,
		CleanForms:   cleanFormsCopy,
		DecodedForms: decodedFormsCopy,
	}
}
