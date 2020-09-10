// words-1a is a simple graph-based program to find word ladders
// between pairs of words in a dictionary. It stores words as nodes
// within the graph, constructing all edges between words on
// addition of the words to the graph.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gonum.org/v1/gonum/graph"
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

	pth := path.DijkstraAllFrom(wg.nodeFor(*first), wg)
	ladders, _ := pth.AllTo(wg.nodeFor(*last).ID())

	for _, l := range ladders {
		fmt.Println(l)
	}
}

// wordGraph is a graph of Hamming distance-1 word paths. It encapsulates
// a Gonum simple.UndirectedGraph to provide a domain-specific API for
// handling word ladder searches.
type wordGraph struct {
	n   int
	ids map[string]int64

	*simple.UndirectedGraph
}

// newWordGraph returns a new wordGraph for words of n characters.
func newWordGraph(n int) wordGraph {
	return wordGraph{
		n:               n,
		ids:             make(map[string]int64),
		UndirectedGraph: simple.NewUndirectedGraph(),
	}
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

	// We know the node is not yet in the graph, so we can add it.
	u := g.UndirectedGraph.NewNode()
	uid := u.ID()
	u = node{word: word, id: uid}
	g.UndirectedGraph.AddNode(u)
	g.ids[word] = uid

	// Join to all the neighbours from words we already know.
	for _, v := range neighbours(word, g.ids) {
		v := g.UndirectedGraph.Node(g.ids[v])
		g.SetEdge(simple.Edge{F: u, T: v})
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
					// We have found a neighbouring word so we
					// can add it to our list of neighbours.
					adj = append(adj, w)
				}
			}
		}
	}
	return adj
}

// nodeFor returns a graph.Node representing the word for inclusion in a wordGraph.
func (g wordGraph) nodeFor(word string) graph.Node {
	id, ok := g.ids[word]
	if !ok {
		return nil
	}
	return g.UndirectedGraph.Node(id)
}

// node is a word node in a wordGraph.
type node struct {
	word string
	id   int64
}

func (n node) ID() int64      { return n.id }
func (n node) String() string { return n.word }
