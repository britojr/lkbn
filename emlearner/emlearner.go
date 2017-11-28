// Package emlearner implements Expectation-Maximization algorithm for parameter estimation
package emlearner

import (
	"log"
	"math"

	"github.com/britojr/lkbn/factor"
	"github.com/britojr/lkbn/inference"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/utl/conv"
)

// property map options
const (
	ParmMaxIters  = "em_max_iters" // maximum number of iterations
	ParmThreshold = "em_threshold" // minimum improvement threshold
	ParmReuse     = "em_use_parms" // use parms of the given model as starting point
)

// default properties
const (
	cMaxIters  = 5
	cThreshold = 1e-1
)

// EMLearner implements Expectation-Maximization algorithm
type EMLearner interface {
	SetProperties(props map[string]string)
	Run(m model.Model, evset []map[int]int) (model.Model, float64, int)
}

// implementation of EMLearner
type emAlg struct {
	maxIters  int     // max number of em iterations
	threshold float64 // minimum improvement threshold
	reuse     bool    // use parms of the given model
	nIters    int     // number of iterations of current alg
}

// New creates a new EMLearner
func New() EMLearner {
	e := new(emAlg)
	// set defaults
	e.maxIters = cMaxIters
	e.threshold = cThreshold
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
	// validate properties
	if e.maxIters <= 0 {
		log.Panicf("emlearner: max iterations (%v) must be > 0", e.maxIters)
	}
	if e.threshold <= 0 {
		log.Panicf("emlearner: convergence threshold (%v) must be > 0", e.threshold)
	}
}

func (e *emAlg) PrintProperties() {
	log.Printf("=========== EM PROPERTIES ========================\n")
	log.Printf("%v: %v\n", ParmMaxIters, e.maxIters)
	log.Printf("%v: %v\n", ParmThreshold, e.threshold)
}

// start defines a starting point for model's parameters
func (e *emAlg) start(infalg inference.InfAlg, evset []map[int]int) {
	// TODO: add a non-trivial em (re)start policy
	// for now, just randomly starts
	if e.reuse {
		return
	}
	for _, nd := range infalg.CTNodes() {
		nd.Potential().RandomDistribute()
	}
}

// Run runs EM until convergence or max iteration number is reached
func (e *emAlg) Run(m model.Model, evset []map[int]int) (model.Model, float64, int) {
	e.PrintProperties()
	log.Printf("emlearner: start\n")
	infalg := inference.NewCTreeCalibration(m)
	e.nIters = 0
	e.start(infalg, evset)
	var llant, llnew float64
	for {
		llnew = e.runStep(infalg, evset)
		e.nIters++
		if llant != 0 {
			if e.nIters >= e.maxIters || (math.Abs((llnew-llant)/llant) < e.threshold) {
				break
			}
			// log.Printf("\temlearner: diff=%v\n", math.Abs((llnew-llant)/llant))
		}
		log.Printf("\temlearner: new=%v\n", llnew)
		llant = llnew
	}
	log.Printf("emlearner: iterations=%v\n", e.nIters)
	return infalg.UpdatedModel(), llnew, e.nIters
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
			panic("emlearner: invalid log(0)")
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
