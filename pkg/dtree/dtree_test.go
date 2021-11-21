package dtree

import (
	"fmt"
	"testing"
)

func printTree(t *Tree, i int) {
	i++
	for b, br := range t.branch {
		fmt.Printf("Level: %d, branch: %s value: %v\n", i, b, br.value)
		printTree(br.t, i)
	}
}

func TestAdd(t *testing.T) {
	fmt.Println("TestAdd")
	tr := &Tree{
		//branch: make(map[string]*Tree),
	}
	if err := tr.Add([]string{}, "foo11"); err != nil {
		t.Error(err)
	}
	//fmt.Println("test1")
	//printTree(tr, 0)
	if err := tr.Add([]string{"a"}, "foo21"); err != nil {
		t.Error(err)
	}
	//fmt.Println("test2")
	//printTree(tr, 0)
	tr = &Tree{
		//branch: make(map[string]*Tree),
	}
	if err := tr.Add([]string{"a"}, "foo22"); err != nil {
		t.Error(err)
	}
	if err := tr.Add([]string{}, "foo12"); err != nil {
		t.Error(err)
	}
	if err := tr.Add([]string{"*", "*"}, "foo31"); err != nil {
		t.Error(err)
	}
	if err := tr.Add([]string{"b", "*", "d", "e"}, "foo41"); err != nil {
		t.Error(err)
	}
	printTree(tr, 0)
}

func TestTreeGetLeafValue(t *testing.T) {
	fmt.Println("TestTreeGetLeafValue")
	tr := &Tree{}
	for x, tt := range []struct {
		path  []string
		value string
	}{
		{[]string{"a", "b"}, "value0"},
		{[]string{"a", "c"}, "value1"},
		{[]string{"a", "d"}, "value2"},
		{[]string{"b"}, "value2"},
		{[]string{"c", "a"}, "value3"},
		{[]string{"c", "b", "a"}, "value4"},
		{[]string{"c", "d"}, "value5"},
	} {
		// Value shouldn't exist before addition.
		if value := tr.GetLeafValue(tt.path); nil != value {
			t.Errorf("#%d: got %v, expected %v", x, value, nil)
		}
		if err := tr.Add(tt.path, tt.value); err != nil {
			t.Error(err)
		}
		value := tr.GetLeafValue(tt.path)
		// Value should exist on successful addition.
		if tt.value != value {
			t.Errorf("#%d: got %v, expected %v", x, value, tt.value)
		}
	}
	//printTree(tr, 0)
}

func TestTreeGetLeafValueWithWildCards(t *testing.T) {
	fmt.Println("TestTreeGetLeafValueWithWildCards")
	tr := &Tree{}
	for _, tt := range []struct {
		path  []string
		value string
	}{
		{[]string{"a", "*"}, "value0"},
		{[]string{"a", "c"}, "value1"},
		{[]string{"a", "b", "*"}, "value2"},
	} {
		if err := tr.Add(tt.path, tt.value); err != nil {
			t.Error(err)
		}
	}
	for x, tt := range []struct {
		path  []string
		value string
	}{
		{[]string{"a", "b"}, "value0"},
		{[]string{"a", "b", "c"}, "value2"},
	} {
		value := tr.GetLeafValue(tt.path)
		// Value should exist on successful addition.
		if tt.value != value {
			t.Errorf("#%d: got %v, expected %v", x, value, tt.value)
		}
	}
	printTree(tr, 0)
}

func TestTreeGetLpm(t *testing.T) {
	fmt.Println("TestTreeGetLpm")
	tr := &Tree{}
	for _, tt := range []struct {
		path  []string
		value string
	}{
		{[]string{"a"}, "a"},
		{[]string{"a", "b", "*"}, "ab*"},
		{[]string{"a", "c", "*"}, "ac*"},
		{[]string{"a", "c", "*", "d", "*"}, "ac*d*"},
		{[]string{"a", "c", "*", "d", "*", "e", "*"}, "ac*d*e*"},
		{[]string{"a", "c", "*", "d", "*", "f", "*"}, "ac*d*f*"},
		{[]string{"a", "c", "*", "d", "*", "g", "*"}, "ac*d*g*"},
	} {
		if err := tr.Add(tt.path, tt.value); err != nil {
			t.Error(err)
		}
	}
	printTree(tr, 0)
	for x, tt := range []struct {
		path  []string
		value interface{}
	}{
		{[]string{"x", "y"}, nil},
		{[]string{"a"}, "a"},
		{[]string{"a", "b", "c"}, "ab*"},
		{[]string{"a", "b", "z"}, "ab*"},
		{[]string{"a", "c", "z"}, "ac*"},
		{[]string{"a", "c", "z", "d", "d", "g", "g", "g", "g"}, "ac*d*g*"},
	} {
		value := tr.GetLpm(tt.path)
		// Value should exist on successful addition.
		if tt.value != value {
			t.Errorf("#%d: got %v, expected %v", x, value, tt.value)
		}
	}
	printTree(tr, 0)
}
