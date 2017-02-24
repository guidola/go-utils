package helper

import (
	"testing"
	"strings"
	"github.com/stretchr/testify/assert"
)

func TestContainsString(t *testing.T) {

	test_cases := []struct{
		in_string string
		search_string string
		result bool
	}{
		{"hola bon dia que tal", "bon", true},
		{"hola bon dia que tal", "bona", false},
		{"hola bon dia que tal", "bo", false},
		{"hola bon dia que tal", "hola bon dia que tal", true},
		{"", "", true},
		{"hola bon dia que tal", "", false},
		{"", "bon", false},
		{"hola bon dia que Tal", "tal", false},
		{"hola bon dia que Tal", "Tal", true},
		{"hola bon dia que Tal!≤", "Tal!≤", true},
		{"hola bon dia que Tal!≤", "Tal!≤ hello", false},
	}


	for _, test_case := range test_cases {
		test_strings := strings.Split(test_case.in_string, " ")
		search_strings := strings.Split(test_case.search_string, " ")
		contains := ContainsString(test_strings, search_strings)
		assert.Equal(t, test_case.result, contains)
	}

}
