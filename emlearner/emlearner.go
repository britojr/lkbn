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
	ParmInitIters = "em_init_iters" // number of intial iterations
	ParmRestarts  = "em_restarts"   // number of starting points
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
	SetProperties(props map[string]string)
	Run(m *model.CTree, evset []map[int]int) (*model.CTree, float64, int)
}

// implementation of EMLearner
type emAlg struct {
	reuse        bool    // use parms of the given model
	threshold    float64 // minimum improvement threshold
	maxIters     int     // max number of em iterations
	numRestarts  int     // number of EM starting points
	numInitIters int     // number of initial iterations

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
}

func (e *emAlg) PrintProperties() {
	log.Printf("=========== EM PROPERTIES ========================\n")
	log.Printf("%v: %v\n", ParmMaxIters, e.maxIters)
	log.Printf("%v: %v\n", ParmThreshold, e.threshold)
	log.Printf("%v: %v\n", ParmReuse, e.reuse)
	log.Printf("%v: %v\n", ParmInitIters, e.numInitIters)
	log.Printf("%v: %v\n", ParmRestarts, e.numRestarts)
}

// start defines a starting point for model's parameters
func (e *emAlg) start(m *model.CTree, evset []map[int]int) inference.InfAlg {
	// set intial candidates
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
	e.PrintProperties()
	e.nIters = 0
	infalg := e.start(m, evset)
	var llant, llnew float64
	for {
		llnew = e.runStep(infalg, evset)
		e.nIters++
		if e.nIters >= e.maxIters || math.Abs(llnew-llant) < e.threshold {
			break
		}
		llant = llnew
		// log.Printf("\temlearner: new=%v\n", llnew)
	}
	// log.Printf("emlearner: iterations=%v\n", e.nIters)
	infalg.CTree().SetScore(llnew)
	return infalg.CTree(), llnew, e.nIters
}

// runStep runs expectation and maximization steps
// returning the loglikelihood of the model with new parameters
func (e *emAlg) runStep(infalg inference.InfAlg, evset []map[int]int) float64 {
	// copy of parameters to hold the sufficient statistics
	count := make(map[*model.CTNode]*factor.Factor)
	var ll float64
	// expecttation step
	for _, evid := range evset {
		evLkhood := infalg.Run(evid)
		// log.Printf("\t>>emlearner: evidlkhood= %v\n", evLkhood)
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

	// maximization step
	for nd, p := range count {
		if pa := nd.Parent(); pa != nil {
			p.Normalize(nd.Potential().Variables().Diff(pa.Potential().Variables())...)
		} else {
			p.Normalize()
		}
		// updates parameters
		nd.SetPotential(p)
	}
	// log.Printf("\t>>emlearner: tot ll= %v\n", ll)
	return ll
}
