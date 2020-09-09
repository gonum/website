// words-2f is a simple graph-based program to find word ladders
// between pairs of words in a dictionary. It stores words as nodes
// within the graph, edges are implied by Hamming distance and are
// enumerated lazily when neighbouring nodes are queried, though
// unlike words-2a, all neighbours are first copied into a slice
// for iteration.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/iterator"
	"gonum.org/v1/gonum/graph/path"
)

func main() {
	first := flag.String("first", "", "first word in word ladder (required - length must match last)")
	last := flag.String("last", "", "last word in word ladder (required - length must match first)")
	flag.Parse()

	if *first == "" || *last == "" || len(*first) != len(*last) {
		flag.Usage()
		os.Exit(2)
	}

	// Make a new word graph and include the first and last
	// words in the ladder in case they do not exists in the
	// dictionary.
	wg := newWordGraph(len(*first))
	for _, p := range []*string{first, last} {
		s := strings.ToLower(*p)
		if !isWord(s) {
			fmt.Fprintf(os.Stderr, "word must not contain punctuation or numerals: %q\n", *p)
			os.Exit(2)
		}
		*p = s
		wg.include(s)
	}

	// Read in a list of unique words from the input stream.
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		wg.include(sc.Text())
	}
	if err := sc.Err(); err != nil {
		log.Fatalf("failed to read word list: %v", err)
	}

	pth := path.DijkstraFrom(wg.nodeFor(*first), wg)
	ladder, _ := pth.To(wg.nodeFor(*last).ID())

	for _, w := range ladder {
		fmt.Println(w)
	}
}

// wordGraph is a graph of Hamming distance-1 word paths using lazy implicit
// edge calculation.
type wordGraph struct {
	n     int
	words []string
	ids   map[string]int64
}

// newWordGraph returns a new wordGraph for words of n characters.
func newWordGraph(n int) wordGraph {
	return wordGraph{n: n, ids: make(map[string]int64)}
}

// include adds word to the graph and connects it to its Hamming distance-1
// neighbours.
func (g *wordGraph) include(word string) {
	if len(word) != g.n || !isWord(word) {
		return
	}
	word = strings.ToLower(word)
	if _, exists := g.ids[word]; exists {
		return
	}
	g.ids[word] = int64(len(g.words))
	g.words = append(g.words, word)
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

// nodeFor returns a graph.Node representing the word for inclusion in a wordGraph.
func (g wordGraph) nodeFor(word string) graph.Node {
	id, ok := g.ids[word]
	if !ok {
		return nil
	}
	return node{word, id}
}

// From implements the graph.Graph From method.
func (g wordGraph) From(id int64) graph.Nodes {
	if uint64(id) >= uint64(len(g.words)) {
		return graph.Empty
	}
	return iterator.NewOrderedNodes(neighbours(g.words[id], g.ids))
}

// neighbours returns a slice of node of words in the words map
// that are within Hamming distance one from the query word.
func neighbours(word string, words map[string]int64) []graph.Node {
	var adj []graph.Node
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
					// We have found a neighbouring word so we
					// can add it to our list of neighbours.
					adj = append(adj, node{word: w, id: words[w]})
				}
			}
		}
	}
	return adj
}

// Edge implements the graph.Graph Edge method.
func (g wordGraph) Edge(uid, vid int64) graph.Edge {
	if uid == vid {
		return nil
	}
	if uint64(uid) >= uint64(len(g.words)) {
		return nil
	}
	if uint64(vid) >= uint64(len(g.words)) {
		return nil
	}
	u := g.words[uid]
	v := g.words[vid]
	d := hamming(u, v)
	if d != 1 {
		return nil
	}
	return edge{f: node{u, uid}, t: node{v, vid}}
}

// hamming returns the Hamming distance between the words a and b.
func hamming(a, b string) int {
	if len(a) != len(b) {
		panic("word length mismatch")
	}
	var d int
	for i, c := range []byte(a) {
		if c != b[i] {
			d++
		}
	}
	return d
}

// node is a word node in a wordGraph.
type node struct {
	word string
	id   int64
}

func (n node) ID() int64      { return n.id }
func (n node) String() string { return n.word }

// edge is a Hamming distance-1 relationship between words in a wordGraph.
type edge struct{ f, t node }

func (e edge) From() graph.Node         { return e.f }
func (e edge) To() graph.Node           { return e.t }
func (e edge) ReversedEdge() graph.Edge { return edge{f: e.t, t: e.f} }
