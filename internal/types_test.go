package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCombine(t *testing.T) {
	l1 := &SwadeshList{
		Group: "1",
		List: []*Word{
			{Group: "1", SwadeshID: 0, Forms: []string{">a"}, CleanForms: []string{"a"}},
			{Group: "1", SwadeshID: 1, Forms: []string{">b"}, CleanForms: []string{"b"}},
			{Group: "1", SwadeshID: 4, Forms: []string{">e"}, CleanForms: []string{"e"}},
		},
	}
	l2 := &SwadeshList{
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
