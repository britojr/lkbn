// Package emlearner implements Expectation-Maximization algorithm for parameter estimation
package emlearner

import (
	"log"
	"math"
	"sort"

	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/inference"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/utl/conv"
)

// property map options
const (
	ParmMaxIters  = "em_max_iters"  // maximum number of iterations
	ParmThreshold = "em_threshold"  // minimum improvement threshold
	ParmReuse     = "em_use_parms"  // use parms of the given model as starting point
	ParmInitIters = "em_init_iters" // number of initial iterations
	ParmRestarts  = "em_restarts"   // number of starting points
	ParmThreads   = "em_threads"    // number of threads for parallelization
)

// default properties
const (
	cMaxIters     = 5
	cThreshold    = 1e-1
	cNumInitIters = 1
	cNumRestarts  = 1
)

// EMLearner implements Expectation-Maximization algorithm
type EMLearner interface {
	Run(m *model.CTree, evset []map[int]int) (*model.CTree, float64, int)
	SetProperties(props map[string]string)
	PrintProperties()
}

// implementation of EMLearner
type emAlg struct {
	reuse        bool    // use parms of the given model
	threshold    float64 // minimum improvement threshold
	maxIters     int     // max number of em iterations
	numRestarts  int     // number of EM starting points
	numInitIters int     // number of initial iterations
	numThreads   int     // number of threads to execute runStep

	nIters int // number of iterations of current alg
}

// New creates a new EMLearner
func New() EMLearner {
	e := new(emAlg)
	// set defaults
	e.threshold = cThreshold
	e.maxIters = cMaxIters
	e.numRestarts = cNumRestarts
	e.numInitIters = cNumInitIters
	return e
}

func (e *emAlg) SetProperties(props map[string]string) {
	// set properties
	if maxIters, ok := props[ParmMaxIters]; ok {
		e.maxIters = conv.Atoi(maxIters)
	}
	if threshold, ok := props[ParmThreshold]; ok {
		e.threshold = conv.Atof(threshold)
	}
	if reuse, ok := props[ParmReuse]; ok {
		e.reuse = conv.Atob(reuse)
	}
	if initIters, ok := props[ParmInitIters]; ok {
		e.numInitIters = conv.Atoi(initIters)
	}
	if restarts, ok := props[ParmRestarts]; ok {
		e.numRestarts = conv.Atoi(restarts)
	}
	if threads, ok := props[ParmThreads]; ok {
		e.numThreads = conv.Atoi(threads)
	}
	// validate properties
	if e.threshold <= 0 {
		log.Panicf("emlearner: convergence threshold (%v) must be > 0", e.threshold)
	}
	if e.maxIters <= 0 {
		log.Panicf("emlearner: max iterations (%v) must be > 0", e.maxIters)
	}
	if e.numRestarts <= 0 {
		log.Panicf("emlearner: num restarts (%v) must be > 0", e.numRestarts)
	}
	if e.numInitIters <= 0 {
		log.Panicf("emlearner: num initial iterations (%v) must be > 0", e.numInitIters)
	}
	if e.numThreads < 0 {
		log.Panicf("emlearner: num threads (%v) cannot be negative", e.numThreads)
	}
}

func (e *emAlg) PrintProperties() {
	log.Printf("=========== EM PROPERTIES ========================\n")
	log.Printf("%v: %v\n", ParmMaxIters, e.maxIters)
	log.Printf("%v: %v\n", ParmThreshold, e.threshold)
	log.Printf("%v: %v\n", ParmReuse, e.reuse)
	log.Printf("%v: %v\n", ParmInitIters, e.numInitIters)
	log.Printf("%v: %v\n", ParmRestarts, e.numRestarts)
	log.Printf("%v: %v\n", ParmThreads, e.numThreads)
}

// start defines a starting point for model's parameters
func (e *emAlg) start(m *model.CTree, evset []map[int]int) inference.InfAlg {
	// set initial candidates
	infalg := make([]inference.InfAlg, e.numRestarts)
	infalg[0] = inference.NewCTreeCalibration(m)
	if !e.reuse {
		for _, nd := range infalg[0].CTNodes() {
			nd.Potential().RandomDistribute()
		}
	}
	for i := 1; i < len(infalg); i++ {
		infalg[i] = inference.NewCTreeCalibration(m.Copy())
		for _, nd := range infalg[i].CTNodes() {
			nd.Potential().RandomDistribute()
		}
	}

	// run initial iterations for each candidate
	for j := 0; j < e.numInitIters; j++ {
		for i := range infalg {
			ll := e.runStep(infalg[i], evset)
			infalg[i].CTree().SetScore(ll)
			// fmt.Printf("ini: %v=%v\n", i, ll)
		}
		e.nIters++
	}

	// start pyramid strategy, eliminates half of the candidates at each iteration
	numItersRound := 1
	for len(infalg) > 1 && e.nIters < e.maxIters {
		for j := 0; j < numItersRound; j++ {
			improved := false
			for i := range infalg {
				ll := e.runStep(infalg[i], evset)
				ct := infalg[i].CTree()
				if ll-ct.Score() > e.threshold {
					improved = true
				}
				ct.SetScore(ll)
				// fmt.Printf("step: %v=%v\n", i, ll)
			}
			e.nIters++
			if !improved {
				return infalg[0]
			}
		}

		// sort in descending scores
		sort.Slice(infalg, func(i int, j int) bool {
			return infalg[i].CTree().Score() > infalg[j].CTree().Score()
		})

		numItersRound *= 2
		if numItersRound > (e.maxIters - e.nIters) {
			numItersRound = e.maxIters - e.nIters
		}
		infalg = infalg[:int(len(infalg)/2)]
	}

	return infalg[0]
}

// Run runs EM until convergence or max iteration number is reached
func (e *emAlg) Run(m *model.CTree, evset []map[int]int) (*model.CTree, float64, int) {
	// e.PrintProperties()
	e.nIters = 0
	infalg := e.start(m, evset)
	var llant, llnew float64
	for {
		llnew = e.runStep(infalg, evset)
		e.nIters++
		if e.nIters >= e.maxIters || math.Abs(llnew-llant) < e.threshold {
			break
		}
		// fmt.Printf("main: %v\t%v\n", llnew, math.Abs(llnew-llant))
		llant = llnew
	}
	infalg.CTree().SetScore(llnew)
	return infalg.CTree(), llnew, e.nIters
}

// runStep runs expectation and maximization steps
// returning the loglikelihood of the model with new parameters
func (e *emAlg) runStep(infalg inference.InfAlg, evset []map[int]int) float64 {
	if e.numThreads <= 1 {
		return e.runStepSequential(infalg, evset)
	}
	return e.runStepParallel(infalg, evset)
}

func (e *emAlg) runStepSequential(infalg inference.InfAlg, evset []map[int]int) float64 {
	// copy of parameters to hold the sufficient statistics
	count, ll := expectStep(infalg, evset)
	maxStep(count)
	return ll
}

// expecttaion step
func expectStep(infalg inference.InfAlg, evset []map[int]int) (map[*model.CTNode]*factor.Factor, float64) {
	// copy of parameters to hold the sufficient statistics
	count := make(map[*model.CTNode]*factor.Factor)
	var ll float64
	for _, evid := range evset {
		evLkhood := infalg.Run(evid)
		if evLkhood == 0 {
			log.Panicf("emlearner: invalid probo of evidence == 0")
		}
		ll += math.Log(evLkhood)

		// acumulate sufficient statistics in the copy of parameters
		for _, nd := range infalg.CTNodes() {
			q, _ := infalg.CalibPotential(nd).Normalize()
			if p, ok := count[nd]; ok {
				p.Plus(q)
			} else {
				count[nd] = q.Copy()
			}
		}
	}
	return count, ll
}

// maximization step
func maxStep(count map[*model.CTNode]*factor.Factor) {
	for nd, p := range count {
		if pa := nd.Parent(); pa != nil {
			p.Normalize(nd.Potential().Variables().Diff(pa.Potential().Variables())...)
		} else {
			p.Normalize()
		}
		// updates parameters
		nd.SetPotential(p)
	}
}

func (e *emAlg) runStepParallel(infalg inference.InfAlg, evset []map[int]int) float64 {
	m := infalg.CTree()
	countCh := make(chan map[*model.CTNode]*factor.Factor, e.numThreads)
	llCh := make(chan float64, e.numThreads)
	size, remain := len(evset)/e.numThreads, len(evset)%e.numThreads
	ini, end := 0, size
	nch := 0
	for ini < len(evset) {
		if remain > 0 {
			remain--
			end++
		}
		go expectStepCh(inference.NewCTreeCalibration(m), evset[ini:end], countCh, llCh)
		nch++
		ini, end = end, end+size
	}

	ll := <-llCh
	count := <-countCh
	for i := 1; i < nch; i++ {
		ll += <-llCh
		for nd, p := range <-countCh {
			count[nd].Plus(p)
		}
	}
	maxStep(count)
	return ll
}

func expectStepCh(
	infalg inference.InfAlg, evset []map[int]int,
	countCh chan<- map[*model.CTNode]*factor.Factor, llCh chan<- float64,
) {
	count, ll := expectStep(infalg, evset)
	countCh <- count
	llCh <- ll
}
