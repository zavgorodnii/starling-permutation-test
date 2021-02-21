package src

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCombine(t *testing.T) {
	l1 := &Wordlist{
		Group: "1",
		List: []*Word{
			{Group: "1", SwadeshID: 0, Forms: []string{">a"}, CleanForms: []string{"a"}},
			{Group: "1", SwadeshID: 1, Forms: []string{">b"}, CleanForms: []string{"b"}},
			{Group: "1", SwadeshID: 4, Forms: []string{">e"}, CleanForms: []string{"e"}},
		},
	}
	l2 := &Wordlist{
		Group: "2",
		List: []*Word{
			{Group: "1", SwadeshID: 1, Forms: []string{">bb"}, CleanForms: []string{"bb"}},
			{Group: "1", SwadeshID: 2, Forms: []string{">cc"}, CleanForms: []string{"cc"}},
			{Group: "1", SwadeshID: 3, Forms: []string{">dd"}, CleanForms: []string{"dd"}},
			{Group: "1", SwadeshID: 4, Forms: []string{">ee"}, CleanForms: []string{"ee"}},
		},
	}
	l3 := l1.Combine(l2)
	if len(l3.List) != 5 {
		t.Fatalf("Combined length is %d, expected 5", len(l3.List))
	}

	assert.Equal(t, []string{">b", ">bb"}, l3.List[1].Forms)
	assert.Equal(t, []string{"b", "bb"}, l3.List[1].CleanForms)

	assert.Equal(t, []string{">e", ">ee"}, l3.List[4].Forms)
	assert.Equal(t, []string{"e", "ee"}, l3.List[4].CleanForms)
}

func TestCompare(t *testing.T) {
	var (
		l1, l2  = getTestWordlists()
		weights = &WeightsStore{
			swadeshIDToWeight: map[int]float64{1: 40., 2: 50, 4: 60},
		}
	)
	var (
		costs1, _ = l1.Compare(l2, weights)
		costs2, _ = l2.Compare(l1, weights)
	)
	assert.Equal(t, costs1, costs2)
}

func getTestWordlists() (l1, l2 *Wordlist) {
	l1 = &Wordlist{
		Group: "1",
		List: []*Word{
			{Group: "1", SwadeshID: 1, DecodedForms: []string{"aaa", "bbbb", "ccccc"}},
			{Group: "1", SwadeshID: 2, DecodedForms: []string{"aaa", "bbbb"}},
			{Group: "1", SwadeshID: 3, DecodedForms: []string{"aaa", "bbbb", "ccccc"}},
			{Group: "1", SwadeshID: 4, DecodedForms: []string{"bbbb", "ccccc"}},
		},
	}
	for _, w := range l1.List {
		w.CleanForms = w.DecodedForms
	}
	l2 = &Wordlist{
		Group: "2",
		List: []*Word{
			{Group: "2", SwadeshID: 1, DecodedForms: []string{"aaa", "bbbb"}},
			{Group: "2", SwadeshID: 2, DecodedForms: []string{"ccccc"}},
			{Group: "2", SwadeshID: 3, DecodedForms: []string{}},
			{Group: "2", SwadeshID: 4, DecodedForms: []string{"aaa"}},
		},
	}
	for _, w := range l2.List {
		w.CleanForms = w.DecodedForms
	}

	return
}
