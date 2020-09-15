// words-3 is a simple graph-based program to find longest word ladders
// between pairs of words in a dictionary. It stores words as nodes
// within the graph, edges are implied by Hamming distance and are
// enumerated lazily when neighbouring nodes are queried.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/iterator"
	"gonum.org/v1/gonum/graph/path"
)

func main() {
	n := flag.Int("n", 0, "length of words to use for ladder (must be greater than 0)")
	flag.Parse()

	if *n <= 0 {
		flag.Usage()
		os.Exit(2)
	}

	// Make a new word graph.
	wg := newWordGraph(*n)

	// Read in a list of unique words from the input stream.
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		wg.include(sc.Text())
	}
	if err := sc.Err(); err != nil {
		log.Fatalf("failed to read word list: %v", err)
	}

	var longest struct {
		length float64
		ends   [][2]int64
	}
	pths := path.DijkstraAllPaths(wg)
	words := graph.NodesOf(wg.Nodes())
	for i, from := range words {
		for _, to := range words[i+1:] {
			fid := from.ID()
			tid := to.ID()
			length := pths.Weight(fid, tid)
			switch {
			case math.IsInf(length, 1):
				continue
			case length > longest.length:
				longest.length = length
				longest.ends = [][2]int64{{fid, tid}}
			case length == longest.length:
				longest.ends = append(longest.ends, [2]int64{fid, tid})
			}
		}
	}
	fmt.Println(longest.length)
	for _, ends := range longest.ends {
		ladders, _ := pths.AllBetween(ends[0], ends[1])
		for _, l := range ladders {
			fmt.Println(l)
		}
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
	return newNeighbours(g.words[id], g.ids)
}

// Edge implements the graph.Graph Edge method.
func (g wordGraph) Edge(uid, vid int64) graph.Edge {
	if !g.HasEdgeBetween(uid, vid) {
		return nil
	}
	return edge{f: node{g.words[uid], uid}, t: node{g.words[vid], vid}}
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

// HasEdgeBetween implements the graph.Undirected HasEdgeBetween method.
func (g wordGraph) HasEdgeBetween(uid, vid int64) bool {
	if uid == vid {
		return false
	}
	if g.Node(uid) == nil || g.Node(vid) == nil {
		return false
	}
	u := g.words[uid]
	v := g.words[vid]
	return hamming(u, v) == 1
}

// Node implements the graph.Graph Node method.
func (g wordGraph) Node(id int64) graph.Node {
	if uint64(id) >= uint64(len(g.words)) {
		return nil
	}
	return node{word: g.words[id], id: id}
}

// Nodes implements the graph.Graph Nodes method.
func (g wordGraph) Nodes() graph.Nodes {
	nodes := make([]graph.Node, len(g.words))
	for w, id := range g.ids {
		nodes[id] = node{word: w, id: id}
	}
	return iterator.NewOrderedNodes(nodes)
}

// neighbours implements the graph.Nodes interface. It is a deterministic
// iterator over sets of nodes that represent words with Hamming distance-1
// from a query word.
type neighbours struct {
	word string
	ids  map[string]int64
	j    int
	d    byte
	buf  []byte
	curr graph.Node
}

// newNeighbours returns a new word neighbours iterator.
func newNeighbours(word string, ids map[string]int64) *neighbours {
	return &neighbours{word: word, ids: ids, d: 'a', buf: make([]byte, len(word))}
}

// Len implements the graph.Nodes Len method. It returns -1 to indicate the iterator
// has an unknown number of of items.
func (it *neighbours) Len() int { return -1 }

// Next implements the graph.Nodes Next method.
func (it *neighbours) Next() bool {
	for it.j < len(it.word) {
		for i, c := range []byte(it.word) {
			if i == it.j {
				it.buf[i] = it.d
			} else {
				it.buf[i] = c
			}
		}
		it.d++
		if it.d > 'z' {
			it.j++
			it.d = 'a'
		}

		if !bytes.Equal(it.buf, []byte(it.word)) {
			// We have found a neighbouring word so we can return
			// true and set the current word to this neighbour.
			if _, ok := it.ids[string(it.buf)]; ok {
				w := string(it.buf)
				it.curr = node{w, it.ids[w]}
				return true
			}
		}
	}
	it.curr = nil
	return false
}

// Node implements the graph.Nodes Node method.
func (it *neighbours) Node() graph.Node { return it.curr }

// Reset implements the graph.Nodes Reset method.
func (it *neighbours) Reset() { it.j, it.d = 0, 'a' }

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
