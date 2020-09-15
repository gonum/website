+++
date = "2020-07-12T12:00:00"
#authors = ["kortschak"] TODO: add authors
draft = false
categories = ["intro", "graph"]
title = "Using Gonum Graphs to Solve Word Ladder Puzzles"
math = true
summary = """
A tour through approaching a simple problem using graphs with Gonum.
"""

[header]
image = ""
caption = ""

+++

<!--
List of word ladder doublets from Martin Gardiner's article.

rogue beast
shoes crust
grass green
black white
costs pence
quell bravo
kettle holder
furies barrel
tears smile
pitch tents
flour bread
raven miser
wheat bread
steal coins
beans shelf
-->

## Word Ladders

[Word ladders](https://books.google.com/books?id=I9oVP8TlyqIC&pg=PA22
), originally [Doublets](https://books.google.com/books?id=JkQCAAAAQAAJ&pg=PP1), is a word game that was invented by Lewis Carrol in 1877. The game involves finding the smallest number of single letter changes required to change a starting word to a target of the same length, only using intermediates that are valid English words. For example, "head" to "tail":

```
HEAD
heal
teal
tell
tall
TAIL
```

An obvious way to resolve this kind of problem mechanically is to represent the word ladder as a path walk through a graph where the nodes are words and edges exist between words with a [Hamming distance](https://en.wikipedia.org/wiki/Hamming_distance) of one. With this representation we can then find the shortest path between the start and end words using something like [Dijkstra's shortest path algorithm](https://en.wikipedia.org/wiki/Dijkstra's_algorithm).

Gonum provides routines for finding shortest paths in the [`graph/path`](https://pkg.go.dev/gonum.org/v1/gonum/graph/path?tab=doc) package. These functions take a [`graph.Graph`](https://pkg.go.dev/gonum.org/v1/gonum/graph?tab=doc#Graph) or [`traverse.Graph`](https://pkg.go.dev/gonum.org/v1/gonum/graph/traverse?tab=doc#Graph) interface type. The use of interface types for representing graphs in Gonum provides a lot of flexibility for representing interesting problems, but also comes with a cost of knowing how to implement the concrete type correctly or how and when to use [the concrete graph types that we provide](https://pkg.go.dev/gonum.org/v1/gonum/graph/simple?tab=doc).

It is worth noting that the simple graph package is included primarily to allow Gonum graph functions to be tested (a multigraph equivalent [`graph/multi`](https://pkg.go.dev/gonum.org/v1/gonum/graph/multi?tab=doc) exists for the same purpose), and that the graph implementations provided may not be the most efficient implementation for any particular graph-based problem; they are reasonably good, generally applicable and relatively simple implementations.

Because we expect that users may need to implement their own graph types that satisfy the `graph` package interfaces we provide [`graph/testgraph`](https://pkg.go.dev/gonum.org/v1/gonum/graph/testgraph?tab=doc), which is a set of testing routines that can be used as a framework to test a graph implementation's adherence to the Gonum graph interface definitions. An example of how this package is used is [here](https://github.com/gonum/gonum/blob/master/graph/simple/directed_test.go). If there is interest, a future post may go through the process of constructing tests for a new graph type using `graph/testgraph`.

With that background out of the way, let's explore some ways that we can use Gonum graphs to solve word ladders in the general case, and then extend the problem to find extreme cases of word ladders, all while examining the performance characteristics of our implementations.

Gonum graphs contain a [`graph.Node`](https://pkg.go.dev/gonum.org/v1/gonum/graph?tab=doc#Node) type whose sole method returns a graph-unique integer ID. So, the naive way to solve the problem would be to read in the word list into a `[]string`, construct a graph based on the Hamming distance relationship that we require and then use the graph node IDs as indexes into the slice. There is a little bit of cleaning up to do since the game does not distinguish letter case and we need to ensure that all the words in our list are the same length.

```
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
	// ... to the last word.
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
					// We have found a neighbouring word so we
					// can add it to our list of neighbours.
					adj = append(adj, w)
				}
			}
		}
	}
	return adj
}
```

Running [this code](/code/word_ladders/words-0/main.go) with Carroll's example gives us an answer in 0.09s, using \~8.5MB. (`xtime` is defined [here](https://blog.golang.org/pprof), `/usr/share/dict/words` is provided by the Ubuntu wamerican package.)
```
$ xtime words-0 -first head -last tail </usr/share/dict/words
head
heal
teal
tell
tall
tail
0.08u 0.00s 0.07r 8516kB words-0 -first head -last tail
```

This implementation certainly works, but the business logic is spread thoughout the input handling and the main loop, so we can clean it up by using the fact that the graph package work with interface values. To do this, we will define a new graph type that holds nodes that are aware of the word that they represent.

The `wordGraph` type still makes use of a Gonum simple graph implementation and the logic of the calculation is essentially identical.

```
type wordGraph struct {
	n   int
	ids map[string]int64

	*simple.UndirectedGraph
}
```

The nodes in the graph hold the word they represent and also satisfy `fmt.Stringer` so that we don't need to do any work to get the word back when we have a path.

```
type node struct {
	word string
	id   int64
}

func (n node) ID() int64      { return n.id }
func (n node) String() string { return n.word }
```

```
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
	n   int // n is the length of words described by the graph.
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
	// The body of neighbours is identical to the version
	// in the original implementations.
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
```
Although the logic is almost identical, the line count is higher, this however comes with a more modular implementation which is easier to understand and would be easier to maintain in a more complex application.

An important aspect of the implementation is that the concrete type wrapping the `*simple.UndirectedGraph` adds domain-specific API that means the client code does not need to deal with the graph, but makes the graph API available to the path-finding function. This is a design intention for the graph packages.

Running [the new version](/code/word_ladders/words-1/main.go) gives us the same performance.
```
$ xtime words-1 -first head -last tail </usr/share/dict/words
head
heal
hell
tell
tall
tail
0.07u 0.01s 0.06r 9008kB words-1 -first head -last tail
```

If you have run the code, you will notice that you won't necessarily get the word ladder matching the one here; there are multiple co-equal shortest paths between "head" and "tail". `gonum/path` provides a way to find all shortest paths between a pair of node which requires only a minor alteration of the code. Instead of calling `path.DijkstraFrom` we'll use [`path.DijkstraAllFrom`](https://pkg.go.dev/gonum.org/v1/gonum/graph/path?tab=doc#DijkstraAllFrom) and loop over the slice of paths returned by [`pth.AllTo`](https://pkg.go.dev/gonum.org/v1/gonum/graph/path?tab=doc#DijkstraAllFrom.AllTo).

```
	pth := path.DijkstraAllFrom(wg.nodeFor(*first), wg)
	ladders, _ := pth.AllTo(wg.nodeFor(*last).ID())

	for _, l := range ladders {
		fmt.Println(l)
	}
```

[This](/code/word_ladders/words-1a/main.go) now gives us all the shortest word ladders from "head" to "tail".
```
$ xtime words-1a -first head -last tail </usr/share/dict/words
[head read reid raid rail tail]
[head heal neal neil nail tail]
[head heal teal tell tall tail]
[head held hell tell tall tail]
[head heal hell tell tall tail]
[head held hell hall tall tail]
[head heal hell hall tall tail]
[head hear heir hair hail tail]
[head held hell hall hail tail]
[head heal hell hall hail tail]
0.07u 0.00s 0.07r 8876kB words-1a -first head -last tail
```

Getting back to the single path problem, the next step is to see what we can do to improve the performance of the program. The observation that can drive this is that when we are constructing the graph in the naive implementation and the `wordGraph` wrapping of that approach, we perform neighbourhood calculations even for words that are unreachable from our doublet. This work will never be used. Depending on the word length and the doublet we choose this can make a reasonable difference.

The main function remains the same, but `wordGraph` is replaced with the following (and goimports adjusts our import list) so that we determine neighbourhoods lazily as we reach them during the shortest path search.

In order to lazily evaluate a word's neighbourhood we will make use of a design feature of Gonum graphs where node and edge iteration is handled by Go interface types. This allows us to replace the built in node iterator with an application-specific iterator that knows more about the nature of the graph we are working with.

To do this, we need to adjust the `wordGraph` a little, adding `From` and `Edge` methods to the type and returning our new node iterator from the `From` method call.

```
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
```

The node iterator itself wraps a slightly altered `neighbours` function used in the previous version with some addition methods required to allow the iterator to be used, implementing the `graph.Nodes` interface.

```
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
	// The Next method is implemented using the same algorithm as used by the
	// neighbours function in the original naive implementation of the search.

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
```

[Here](/code/word_ladders/words-2/main.go) we see a good improvement.
```
$ xtime words-2 -first head -last tail </usr/share/dict/words
head
heal
neal
neil
nail
tail
0.04u 0.00s 0.05r 6472kB words-2 -first head -last tail
```

Note that unlike the previous implementation for a single shortest path, this implementation will always output the same result since the `*neighbours` iterator here is completely deterministic.

A [simpler version of this](/code/word_ladders/words-2f/main.go) exists, that calculates a slice of neighbours for each `From` call,
```
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
					adj = append(adj, node{word: w, id: words[w]})
				}
			}
		}
	}
	return adj
}
```
but has worse performance characteristics.
```
$ xtime words-2f -first head -last tail </usr/share/dict/words >/dev/null
0.07u 0.00s 0.07r 7276kB words-2f -first head -last tail
```

[Again](/code/word_ladders/words-2a/main.go), we can obtain all possible ladders for the doublet with the change to `path.DijkstraAllFrom` with performance that still beats the naive implementation.

```
$ xtime words-2a -first head -last tail </usr/share/dict/words >/dev/null
0.05u 0.00s 0.04r 6964kB words-2a -first head -last tail
```

So it seems that we have an answer, we should use the lazy neighbourhood approach.

We can extend our problem though to not just find one or all solutions for a doublet given a word list, but find the longest ladder for a given word length in a word list. We can do this by replacing the `path.DijkstraFrom` search stanza with the following code:
```
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
```
and replacing the flags we accept.
```
	n := flag.Int("n", 0, "length of words to use for ladder (must be greater than 0)")
```

This change can be made to either the implicit lazy implementation or the eager approach.

However, the [lazy approach](/code/word_ladders/words-3/main.go)
```
$ xtime words-3 -n 4 </usr/share/dict/words
19
[inca inch itch etch each bach bath oath oats opts opus onus anus ants ante anne acne ache achy ashy]
[inca inch itch etch each mach math oath oats opts opus onus anus ants ante anne acne ache achy ashy]
[inca inch itch etch each bach bath bats oats opts opus onus anus ants ante anne acne ache achy ashy]
[inca inch itch etch each mach math mats oats opts opus onus anus ants ante anne acne ache achy ashy]
[inca inch itch etch each bach bath oath oats opts opus onus anus ants ante anne acne ache ashe ashy]
[inca inch itch etch each mach math oath oats opts opus onus anus ants ante anne acne ache ashe ashy]
[inca inch itch etch each bach bath bats oats opts opus onus anus ants ante anne acne ache ashe ashy]
[inca inch itch etch each mach math mats oats opts opus onus anus ants ante anne acne ache ashe ashy]
[chum chug thug thud thad chad clad clan alan alas alms elms elma erma erna edna edda eddy edgy edge]
[chum chug thug thud thad chad chan clan alan alas alms elms elma erma erna edna edda eddy edgy edge]
[chum chug thug thud thad than chan clan alan alas alms elms elma erma erna edna edda eddy edgy edge]
[chum chug thug thud thad than khan klan alan alas alms elms elma erma erna edna edda eddy edgy edge]
[chum chug thug thud thad chad clad clan alan alas alms alma elma erma erna edna edda eddy edgy edge]
[chum chug thug thud thad chad chan clan alan alas alms alma elma erma erna edna edda eddy edgy edge]
[chum chug thug thud thad than chan clan alan alas alms alma elma erma erna edna edda eddy edgy edge]
[chum chug thug thud thad than khan klan alan alas alms alma elma erma erna edna edda eddy edgy edge]
[edge edgy eddy edda edna erna erma elma elms alms aims sims sums sues suet suit quit quid quad quay]
[edge edgy eddy edda edna erna erma elma alma alms aims sims sums sues suet suit quit quid quad quay]
87.07u 0.44s 75.44r 947092kB words-3 -n 4
```
is handily beaten by the [eager approach](/code/word_ladders/words-4/main.go).
```
$ xtime words-4 -n 4 </usr/share/dict/words >/dev/null
28.22u 0.34s 23.36r 863496kB words-4 -n 4
```

This is because the lazy optimisation depends on only expanding neighbourhoods once and avoiding parts that are not accessible from the given end points, but the all pairs shortest paths algorithm repeatedly expands neighbourhoods and examines the entire graph, so it makes sense to do work up front and retain the results.

The learning we can take from this is that one particular graphical approach that suits one particular problem may be a poor fit for another, even closely related, problem. So an understanding of the problem that you are attempting to solve and at least a passing understanding of the algorithmic details of the functions that you plan to use are crucial for being able to choose an appropriate graph implementation. As an aside, the diversity of problems that graphs can be used to address has been one of the biggest challenges in designing the graph packages' API.

For additional fun, an extension that we can look into is the widest doublet, that is the doublet with the greatest number of solutions.

Again, the [lazy implementation](/code/word_ladders/words-5/main.go) is beaten by the [eager naive implementation](/code/word_ladders/words-6/main.go).
```
$ xtime words-5 -n 4 </usr/share/dict/words >/dev/null
769.18u 2.34s 210.60r 2170600kB words-5 -n 4
$ xtime words-6 -n 4 </usr/share/dict/words >/dev/null
663.70u 2.38s 157.56r 2059916kB words-6 -n 4
```

For the record, the doublet here is "stow" and "dave", with 419 solutions.

The full code for each of the word ladder programs is available from the links in the text or by using `go get github.com/gonum/website/static/code/word_ladders/...`. It depends on Gonum version 0.8.1 which added the `path.DijkstraAllFrom` function.

*By Dan Kortschak*
