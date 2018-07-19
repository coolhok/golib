package trie

import (
	"bufio"
	"log"
	"os"
	"sort"
	"testing"
)

func addFromFile(t *Trie, path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewScanner(file)

	for reader.Scan() {
		t.Add(reader.Text(), nil)
	}

	if reader.Err() != nil {
		log.Fatal(err)
	}
}

func TestTrieAdd(t *testing.T) {
	trie := NewTrie()

	n := trie.Add("foo", 1)

	if n.Meta().(int) != 1 {
		t.Errorf("Expected 1, got: %d", n.Meta().(int))
	}
}

func TestTrieFind(t *testing.T) {
	trie := NewTrie()
	trie.Add("foo", 1)

	n, err := trie.Find("foo")
	if err != nil {
		t.Fatal("Could not find node")
	}

	if n.Meta().(int) != 1 {
		t.Errorf("Expected 1, got: %d", n.Meta().(int))
	}
}

func TestRemove(t *testing.T) {
	trie := NewTrie()
	initial := []string{"football", "foostar", "foosball"}

	for _, key := range initial {
		trie.Add(key, nil)
	}

	trie.Remove("foosball")
	keys := trie.Keys()

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys got %d", len(keys))
	}

	for _, k := range keys {
		if k != "football" && k != "foostar" {
			t.Errorf("key was: %s", k)
		}
	}

	keys = trie.FuzzySearch("foo")

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys got %d", len(keys))
	}

	for _, k := range keys {
		if k != "football" && k != "foostar" {
			t.Errorf("Expected football got: %#v", k)
		}
	}
}

func TestTrieKeys(t *testing.T) {
	trie := NewTrie()
	expected := []string{"bar", "foo"}

	for _, key := range expected {
		trie.Add(key, nil)
	}

	kl := len(trie.Keys())
	if kl != 2 {
		t.Errorf("Expected 2 keys, got %d, keys were: %v", kl, trie.Keys())
	}

	keys := trie.Keys()

	sort.Strings(keys)
	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("Expected %#v, got %#v", expected[i], key)
		}
	}
}

func TestPrefixSearch(t *testing.T) {
	trie := NewTrie()
	expected := []string{
		"foo",
		"foosball",
		"football",
		"foreboding",
		"forementioned",
		"foretold",
		"foreverandeverandeverandever",
		"forbidden",
	}

	defer func() {
		r := recover()
		if r != nil {
			t.Error(r)
		}
	}()

	trie.Add("bar", nil)
	for _, key := range expected {
		trie.Add(key, nil)
	}

	tests := []struct {
		pre      string
		expected []string
		length   int
	}{
		{"fo", expected, len(expected)},
		{"foosbal", []string{"foosball"}, 1},
		{"abc", []string{}, 0},
	}

	for _, test := range tests {
		actual := trie.PrefixSearch(test.pre)
		sort.Strings(actual)
		sort.Strings(test.expected)
		if len(actual) != test.length {
			t.Errorf("Expected len(actual) to == %d for pre %s", test.length, test.pre)
		}

		for i, key := range actual {
			if key != test.expected[i] {
				t.Errorf("Expected %v got: %v", test.expected[i], key)
			}
		}
	}

	trie.PrefixSearch("fsfsdfasdf")
}

func TestFuzzySearch(t *testing.T) {
	trie := NewTrie()
	setup := []string{
		"foosball",
		"football",
		"bmerica",
		"ked",
		"kedlock",
		"frosty",
		"bfrza",
		"foo/bart/baz.go",
	}
	tests := []struct {
		partial string
		length  int
	}{
		{"fsb", 1},
		{"footbal", 1},
		{"football", 1},
		{"fs", 2},
		{"oos", 1},
		{"kl", 1},
		{"ft", 3},
		{"fy", 1},
		{"fz", 2},
		{"a", 5},
	}

	for _, key := range setup {
		trie.Add(key, nil)
	}

	for _, test := range tests {
		actual := trie.FuzzySearch(test.partial)

		if len(actual) != test.length {
			t.Errorf("Expected len(actual) to == %d, was %d for %s", test.length, len(actual), test.partial)
		}
	}
}

func TestFuzzySearchSorting(t *testing.T) {
	trie := NewTrie()
	setup := []string{
		"foosball",
		"football",
		"bmerica",
		"ked",
		"kedlock",
		"frosty",
		"bfrza",
		"foo/bart/baz.go",
	}

	for _, key := range setup {
		trie.Add(key, nil)
	}

	actual := trie.FuzzySearch("fz")
	expected := []string{"bfrza", "foo/bart/baz.go"}

	for i, v := range expected {
		if actual[i] != v {
			t.Errorf("Expected %s got %s", v, actual[i])
		}
	}

}

func BenchmarkTieKeys(b *testing.B) {
	trie := NewTrie()
	keys := []string{"bar", "foo", "baz", "bur", "zum", "burzum", "bark", "barcelona", "football", "foosball", "footlocker"}

	for _, key := range keys {
		trie.Add(key, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trie.Keys()
	}
}

func BenchmarkPrefixSearch(b *testing.B) {
	trie := NewTrie()
	addFromFile(trie, "/usr/share/dict/words")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = trie.PrefixSearch("fo")
	}
}

func BenchmarkFuzzySearch(b *testing.B) {
	trie := NewTrie()
	addFromFile(trie, "/usr/share/dict/words")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = trie.FuzzySearch("fs")
	}
}
