package sukdf

import (
	"fmt"
	"math/rand"
	"math"
	"compress/zlib"
	"bytes"
	"io"
)

//go:generate easytags $GOFILE

const N = 16

type Sukdf struct {
	pass          string
	grid          [][]int
	numbers       []int
	rtag          [][]bool
	ctag          [][]bool
	btag          [][]bool
	r             *rand.Rand
	btrackcount   int
	MAX_BACKTRACK int
	Verbose       bool
}

func New(p string, fs ... func(*Sukdf)) (*Sukdf) {
	s := &Sukdf{pass: p, MAX_BACKTRACK: 20000}
	if fs != nil && len(fs) > 0 {
		for _, f := range fs {
			f(s)
		}
	}
	s.grid = make([][]int, N)
	for i := range s.grid {
		s.grid[i] = make([]int, N)
	}
	return s
}

func WithMaxBacktracks(maxbt int) func(*Sukdf) {
	return func(s *Sukdf) { s.MAX_BACKTRACK = maxbt }
}

func (s *Sukdf) Reset() {
	s = New(s.pass)
}

func (s *Sukdf) Compute() ([]byte, bool) {
	s.rtag, s.ctag, s.btag = make([][]bool, N), make([][]bool, N), make([][]bool, N)

	for i, _ := range s.grid {
		s.rtag[i] = make([]bool, N)
		s.ctag[i] = make([]bool, N)
		s.btag[i] = make([]bool, N)
	}

	salt := int64(0)

	for _, c := range s.pass {
		c = c + 1
		n1 := int(c & 0x0f)
		n2 := int((c & 0xf0) >> 4)
		salt += int64(n1) + int64(n2)
		s.addnumberunique(n1)
		s.addnumberunique(n2)
	}

	for i := 1; i < (N + 1); i++ {
		s.addnumberunique(i)
	}

	source := rand.NewSource(salt)
	s.r = rand.New(source)

	rand.Seed(salt)
	s.btrackcount = 0
	success := s.backtrack()
	if !success {
		if s.btrackcount >= s.MAX_BACKTRACK && s.Verbose {
			fmt.Println("backtrack limit reached")
		}
		return getcompressedbuffer(s.grid).Bytes(), success
	}
	s.btrackcount = 0
	completeGrid := s.clonegrid()

	puzzle := s.createPuzzle()
	s.copygrid(completeGrid)

	return getcompressedbuffer(puzzle).Bytes(), success
}

func getcompressedbuffer(puzzle [][]int) *bytes.Buffer {
	parr := []byte{}
	for i := range puzzle {
		for j := range puzzle[i] {
			parr = append(parr, byte(puzzle[i][j]))
		}

	}
	buf := new(bytes.Buffer)
	w := zlib.NewWriter(buf)
	io.Copy(w, bytes.NewReader(parr))
	err := w.Close()
	if err != nil {
		fmt.Println(err)
	}
	return buf
}

func printgrid(g [][]int, name string) {
	fmt.Println(name)
	for i := range g {
		if i%4 == 0 {
			fmt.Println()
		}
		fmt.Printf("%2v   %2v   %2v   %2v   \n", g[i][:4], g[i][4:8], g[i][8:12], g[i][12:])
	}
	fmt.Println()
}

func matchgrids(g1, g2 [][]int) bool {
	for i := range g1 {
		for j := range g2 {
			if g1[i][j] != g2[i][j] {
				return false
			}
		}
	}
	return true
}

func (s *Sukdf) copygrid(g [][]int) {
	for i := range g {
		for j := range g[i] {
			if g[i][j] == 0 && s.grid[i][j] != 0 {
				s.removeNumber(s.grid[i][j], i, j)
			} else if g[i][j] != 0 {
				s.addNumber(g[i][j], i, j)
			}
		}
	}
}

func (s *Sukdf) clonegrid() [][]int {
	clone := make([][]int, N)
	for i := range s.grid {
		clone[i] = make([]int, N)
		for j := range s.grid[i] {
			clone[i][j] = s.grid[i][j]
		}
	}
	return clone
}

func (s *Sukdf) createPuzzle() [][]int {
	ogrid := s.clonegrid()
	status := false
	puzzle := make([][]int, N)
	for i := range puzzle {
		puzzle[i] = make([]int, N)
	}

	for count := 0; count < 100; count ++ {
		n := s.r.Int()%20 + 130
		perm := s.r.Perm(N * N)[:(N*N - n)]
		for _, i := range perm {
			s.removeNumber(s.grid[i/N][i%N], i/N, i%N)
		}
		for i := range s.grid {
			for j := range s.grid[i] {
				puzzle[i][j] = s.grid[i][j]
			}
		}
		for i := 0; i < N; i++ {
			s.copygrid(puzzle)
			status = s.backtrack()
			if !status {
				break
			}
			if !matchgrids(ogrid, s.grid) {
				status = false;
				break
			} else {
				status = true
			}
		}
		if status {
			break
		} else {
			s.copygrid(ogrid)
		}
	}
	if !status {
		s.copygrid(ogrid)
		return ogrid
	}
	return puzzle
}

func (s *Sukdf) addnumberunique(i int) {
	if i == 0 {
		return
	}

	for _, n := range s.numbers {
		if n == i {
			return
		}
	}
	s.numbers = append(s.numbers, i)
}

func blocknumber(i, j int) int {
	base := int(math.Floor(math.Sqrt(float64(N))))
	return base*(i/base) + j/base
}

func (s *Sukdf) isValid(n, i, j int) bool {
	n = n - 1
	return !(s.rtag[i][n] || s.ctag[j][n] || s.btag[blocknumber(i, j)][n]) && s.grid[i][j] == 0
}

func (s *Sukdf) addNumber(n, i, j int) {
	s.grid[i][j] = n
	n = n - 1
	s.btag[blocknumber(i, j)][n] = true
	s.rtag[i][n] = true
	s.ctag[j][n] = true
}

func (s *Sukdf) removeNumber(n, i, j int) {
	s.grid[i][j] = 0
	n = n - 1
	s.btag[blocknumber(i, j)][n] = false
	s.rtag[i][n] = false
	s.ctag[j][n] = false
}

func (s *Sukdf) printgrid() {
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			fmt.Printf("%d ", s.grid[i][j])
		}
		fmt.Println("")
	}
}

func (s *Sukdf) fillSingles() bool {
	var n int
	added := false
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			n = s.grid[i][j]
			if n > 0 {
				continue
			}
			count := 0
			for k := 1; k < (N + 1); k++ {
				if s.isValid(k, i, j) {
					count ++
					n = k
					if count > 1 {
						break
					}
				}
			}
			if count == 1 {
				added = true
				s.addNumber(n, i, j)
			}
		}
	}
	return added
}

func (s *Sukdf) findEmpty() (int, int, bool) {
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			if s.grid[i][j] == 0 {
				return i, j, true
			}
		}
	}
	return -1, -1, false
}

func (s *Sukdf) shuffle() {
	perm := s.r.Perm(N)
	var i, v int
	for i, v = range perm {
		s.numbers[v], s.numbers[i] = s.numbers[i], s.numbers[v]
	}
}

func (s *Sukdf) backtrack() bool {
	r, c, empty := s.findEmpty()
	if !empty {
		return true
	}
	if s.btrackcount > s.MAX_BACKTRACK {
		return false
	}
	s.shuffle()
	numbers := s.numbers

	for i := 0; i < len(numbers); i++ {
		n := numbers[i]
		if s.isValid(n, r, c) {
			s.addNumber(n, r, c)
			ga := s.solveLogical()
			s.btrackcount++
			if s.backtrack() {
				return true
			}
			for _, g := range ga {
				s.removeNumber(g.n, g.i, g.j)
			}
			s.removeNumber(n, r, c)
		}
	}
	return false
}

type guess struct {
	i int
	j int
	n int
}

func (s *Sukdf) solveLogical() []guess {
	var ga []guess
	found := false
	var n int
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			count := 0
			var no int
			for n = 1; n < (N + 1); n++ {
				if s.isValid(n, i, j) {
					count++
					no = n
				}
				if count > 1 {
					break
				}
			}
			if count == 1 {
				s.addNumber(no, i, j)
				found = true
				ga = append(ga, guess{i: i, j: j, n: no})
			}
		}
	}

	if found {
		ga = append(ga, s.solveLogical()...)
	}
	return ga
}
