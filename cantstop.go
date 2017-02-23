package cantstop

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"strings"
)

// Sim collects statistics of a simulation of how often you may retry a configuration of three lanes.
type Sim struct {
	N     int // repetitions
	Stats []*Stats
}

// Stats collects statistics of a simulation of how often you may retry a configuration of three lanes.
type Stats struct {
	Config Config // lane configuration

	e int64 // expected value E = e / n

	// Knuth's sd algorithm
	n              int64
	meanPrev, mean float64
	sPrev, s       float64

	histogram [20]int

	// Use separate rands in a multi-threaded app.
	// Avoid rand.Intn etc. these delegate to a global thread-safe (aka blocking) rand.
	rand *rand.Rand
}

// Config is the selection of three lanes in a game of "can't stop".
type Config []int

const (
	// HistHeight height of a histogram diagram
	HistHeight = 80
)

var (
	// Configs 21+
	// 234 235 236 237 245 246 247 256 257 267 345 346 347 356 357 367 456 457 467 567
	Configs = []Config{
		{2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}, {11}, {12},
		{2, 3, 4}, {2, 3, 5}, {2, 3, 6}, {2, 3, 7}, {2, 4, 5}, {2, 4, 6}, {2, 4, 7}, {2, 5, 6}, {2, 5, 7}, {2, 6, 7},
		{3, 4, 5}, {3, 4, 6}, {3, 4, 7}, {3, 5, 6}, {3, 5, 7}, {3, 6, 7},
		{4, 5, 6}, {4, 5, 7}, {4, 6, 7},
		{5, 6, 7},
		{6, 7, 8},
		{7, 8, 9},
		{10, 11, 12},
	}
)

// NewSim create a simulation with N repetitions.
func NewSim(N int) *Sim {

	sim := new(Sim)
	sim.N = N

	sim.Stats = make([]*Stats, len(Configs))
	for i, cnf := range Configs {
		sim.Stats[i] = &Stats{Config: cnf, rand: rand.New(rand.NewSource(12))}
	}

	return sim
}

// Run the simulation.
func (sim *Sim) Run() *Sim {

	finished := make(chan bool, 200)

	for _, st := range sim.Stats {
		go runStats(sim.N, st, finished)
	}

	for i := 0; i < len(sim.Stats); i++ {
		<-finished
	}

	return sim
}

func runStats(N int, st *Stats, finished chan bool) {

	for i := 0; i < N; i++ {

		tries := 0
		for st.Config.Matches(st.Roll()) {
			tries++
		}

		st.Val(tries)
	}

	finished <- true
}

// Roll four six-dice.
func (st *Stats) Roll() (d1, d2, d3, d4 int) {
	d1 = st.rand.Intn(6) + 1
	d2 = st.rand.Intn(6) + 1
	d3 = st.rand.Intn(6) + 1
	d4 = st.rand.Intn(6) + 1
	return
}

// Matches if a sum of two out of four dice hits a current lane.
func (cnf Config) Matches(d1, d2, d3, d4 int) bool {
	for _, c := range cnf {
		if d1+d2 == c {
			return true
		}
		if d1+d3 == c {
			return true
		}
		if d1+d4 == c {
			return true
		}
		if d2+d3 == c {
			return true
		}
		if d2+d4 == c {
			return true
		}
		if d3+d4 == c {
			return true
		}
	}
	return false
}

// Val adds a new value of successful tries.
func (st *Stats) Val(v int) {

	// Histogram of counts per tries
	if v < len(st.histogram) {
		st.histogram[v]++
	}

	// Expected value
	st.e += int64(v)
	d := float64(v)

	// Standard deviation
	st.n++
	if st.n == 1 {
		st.mean = d
		st.s = 0
	} else {
		st.meanPrev = st.mean
		st.sPrev = st.s

		st.mean = st.meanPrev + (d-st.meanPrev)/float64(st.n)
		st.s = st.sPrev + (d-st.meanPrev)*(d-st.mean)
	}
}

// E is the expected value.
func (st *Stats) E() float64 {
	return float64(st.e) / float64(st.n)
}

// Sd is the standard deviation.
func (st *Stats) Sd() float64 {
	return math.Sqrt(st.s / (float64(st.n) - 1))
}

func (sim *Sim) String() string {

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprint("Lanes      E   Sd       E (Bar)\n"))
	buf.WriteString(fmt.Sprint("----------------------------------------------------------\n"))
	for _, st := range sim.Stats {
		buf.WriteString(fmt.Sprintf("%2v", st.Config))
		buf.WriteString(fmt.Sprintf(" %4.1f", st.E()))
		buf.WriteString(fmt.Sprintf(" %4.1f  \t", st.Sd()))
		buf.WriteString(strings.Repeat("■", int(st.E()*5)))
		buf.WriteString("\n")
	}

	buf.WriteString("\n")

	scale := maxHist(sim) / HistHeight
	for _, st := range sim.Stats {
		buf.WriteString(fmt.Sprintf("%v\n", st.Config))

		for i, h := range st.histogram {
			buf.WriteString(fmt.Sprintf("%2d ", i))
			buf.WriteString(strings.Repeat("■", h/scale))
			buf.WriteString(fmt.Sprintf(" %d\n", h/scale))
			if h/scale == 0 && i != 0 {
				break
			}
		}
	}

	return buf.String()
}

func maxHist(sim *Sim) int {

	max := 0
	for _, st := range sim.Stats {
		for _, h := range st.histogram {
			if h > max {
				max = h
			}
		}
	}

	return max
}
