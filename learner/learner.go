package learner

import (
	"fmt"
	"log"
	"time"

	"github.com/britojr/lkbn/data"
)

// search algorithms names
const (
	AlgCTSampleSearch = "ctsample"
	AlgCTBridgeSearch = "ctbridge"
	AlgBIextSearch    = "biext"
)

// file parameters fields
const (
	ParmTreewidth    = "treewidth"     // structure treewidth
	ParmNumLatent    = "num_latent"    // number of latent variables
	ParmLatentVars   = "latent_vars"   // latent variables cardinalities
	ParmParentScores = "parent_scores" // precomputed parent scores file
	ParmInputNet     = "input_net"     // tree network previously learned

	ParmMaxTimeCluster = "max_time_per_cluster"
	ParmMaxIterCluster = "max_iter_per_cluster"
	ParmCardReduction  = "card_reduction"
)

// Learner defines a structure learner algorithm
type Learner interface {
	Search() Solution
	SetDefaultParameters()
	SetDataset(*data.Dataset)
	SetFileParameters(map[string]string)
	ValidateParameters()
	PrintParameters()
	Treewidth() int
}

// Solution defines a solution interface
type Solution interface {
	Score() float64
}

// Create creates a structure learner algorithm
func Create(learnerAlg string) (lr Learner) {
	creators := map[string]func() Learner{
		AlgCTSampleSearch: NewCTSampleSearch,
		AlgCTBridgeSearch: NewCTBridgeSearch,
		AlgBIextSearch:    NewBIextSearch,
	}
	if create, ok := creators[learnerAlg]; ok {
		lr = create()
		lr.SetDefaultParameters()
		return
	}
	panic(fmt.Errorf("invalid algorithm option: '%v'", learnerAlg))
}

// Search applies the strategy to find the best solution
func Search(alg Learner, numSolutions, timeAvailable int) (Solution, int) {
	var best, current Solution
	if numSolutions <= 0 && timeAvailable <= 0 {
		numSolutions = 1
	}
	i := 0
	if timeAvailable > 0 {
		// TODO: the documentation recommends using NewTimer(d).C instead of time.After
		remaining := time.Duration(timeAvailable) * time.Second
		for {
			ch := make(chan Solution, 1)
			start := time.Now()
			go func() {
				ch <- alg.Search()
			}()
			select {
			case current = <-ch:
				remaining -= time.Since(start)
			case <-time.After(remaining):
				remaining = 0
			}

			if best == nil || current.Score() > best.Score() {
				best = current
			}
			if remaining <= 0 {
				log.Printf("Time out in %v iterations\n", i)
				break
			}
			i++
			if numSolutions > 0 && i >= numSolutions {
				break
			}
		}
	} else {
		for ; i < numSolutions; i++ {
			current := alg.Search()
			if best == nil || current.Score() > best.Score() {
				best = current
			}
		}
	}
	return best, i
}
