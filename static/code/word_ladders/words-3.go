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

	wg := newWordGraph(*n)
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

type wordGraph struct {
	n     int
	words []string
	ids   map[string]int64
}

func newWordGraph(n int) wordGraph {
	return wordGraph{n: n, ids: make(map[string]int64)}
}

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

func isWord(s string) bool {
	for _, c := range []byte(s) {
		if lc(c) < 'a' || 'z' < lc(c) {
			return false
		}
	}
	return true
}

func lc(b byte) byte {
	return b | 0x20
}

func (g wordGraph) nodeFor(word string) graph.Node {
	id, ok := g.ids[word]
	if !ok {
		return nil
	}
	return node{word, id}
}

func (g wordGraph) From(id int64) graph.Nodes {
	if uint64(id) >= uint64(len(g.words)) {
		return graph.Empty
	}
	return newNeighbours(g.words[id], g.ids)
}

func (g wordGraph) Edge(uid, vid int64) graph.Edge {
	if !g.HasEdgeBetween(uid, vid) {
		return nil
	}
	return edge{f: node{g.words[uid], uid}, t: node{g.words[vid], vid}}
}

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

func (g wordGraph) Node(id int64) graph.Node {
	if uint64(id) >= uint64(len(g.words)) {
		return nil
	}
	return node{word: g.words[id], id: id}
}

func (g wordGraph) Nodes() graph.Nodes {
	nodes := make([]graph.Node, len(g.words))
	for w, id := range g.ids {
		nodes[id] = node{word: w, id: id}
	}
	return iterator.NewOrderedNodes(nodes)
}

type neighbours struct {
	word string
	ids  map[string]int64
	j    int
	d    byte
	buf  []byte
	curr graph.Node
}

func newNeighbours(word string, ids map[string]int64) *neighbours {
	return &neighbours{word: word, ids: ids, d: 'a', buf: make([]byte, len(word))}
}

func (it *neighbours) Len() int { return -1 }

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

func (it *neighbours) Node() graph.Node { return it.curr }

func (it *neighbours) Reset() { it.j, it.d = 0, 'a' }

type node struct {
	word string
	id   int64
}

func (n node) ID() int64      { return n.id }
func (n node) String() string { return n.word }

type edge struct{ f, t node }

func (e edge) From() graph.Node         { return e.f }
func (e edge) To() graph.Node           { return e.t }
func (e edge) ReversedEdge() graph.Edge { return edge{f: e.t, t: e.f} }
