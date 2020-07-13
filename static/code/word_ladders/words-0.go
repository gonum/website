// words-0 is a simple graph-based program to find word ladders
// between pairs of words in a dictionary. It uses graph node IDs
// as indexes into the dictionary slice.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

func main() {
	first := flag.String("first", "", "first word in word ladder (required - length must match last)")
	last := flag.String("last", "", "last word in word ladder (required - length must match first)")
	flag.Parse()

	if *first == "" || *last == "" || len(*first) != len(*last) {
		flag.Usage()
		os.Exit(2)
	}
	for _, p := range []*string{first, last} {
		s := strings.ToLower(*p)
		if !isWord(s) {
			fmt.Fprintf(os.Stderr, "word must not contain punctuation or numerals: %q\n", *p)
			os.Exit(2)
		}
		*p = s
	}

	// Read in a list of unique words from the input stream.
	// Include the first and last words in the ladder in case
	// they do not exists in the dictionary.
	words := map[string]int64{*first: 0, *last: 1}
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		w := sc.Text()
		if len(w) != len(*first) || !isWord(w) {
			continue
		}
		w = strings.ToLower(w)
		if _, exists := words[w]; exists {
			continue
		}
		words[w] = int64(len(words))
	}
	if err := sc.Err(); err != nil {
		log.Fatalf("failed to read word list: %v", err)
	}
	list := make([]string, len(words))
	for w, id := range words {
		list[id] = w
	}

	// Construct a graph using Hamming distance one edges from
	// list of words.
	g := simple.NewUndirectedGraph()
	for u, uid := range words {
		for _, v := range neighbours(u, words) {
			vid := words[v]
			g.SetEdge(simple.Edge{F: simple.Node(uid), T: simple.Node(vid)})
		}
	}

	// Find the shortest paths from the first word...
	pth := path.DijkstraFrom(simple.Node(words[*first]), g)
	// ,,, to the last word.
	ladder, _ := pth.To(words[strings.ToLower(*last)])

	// Print each step in the ladder.
	for _, w := range ladder {
		fmt.Println(list[w.ID()])
	}
}

// isWord returns whether s is entirely alphabetical.
func isWord(s string) bool {
	for _, c := range []byte(s) {
		if lc(c) < 'a' || 'z' < lc(c) {
			return false
		}
	}
	return true
}

// lc returns the lower case of b.
func lc(b byte) byte {
	return b | 0x20
}

// neighbours returns a slice of string of words in the words map
// that are within Hamming distance one from the query word.
func neighbours(word string, words map[string]int64) []string {
	var adj []string
	for j := range word {
		for d := byte('a'); d <= 'z'; d++ {
			b := make([]byte, len(word))
			for i, c := range []byte(word) {
				if i == j {
					b[i] = d
				} else {
					b[i] = c
				}
			}
			w := string(b)
			if w != word {
				if _, ok := words[w]; ok {
					adj = append(adj, w)
				}
			}
		}
	}
	return adj
}
