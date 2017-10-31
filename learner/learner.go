package learner

import (
	"fmt"
	"log"
	"time"

	"github.com/britojr/lkbn/data"
)

// search algorithms names
const (
	AlgSampleSearch = "sample"
)

// file parameters fields
const (
	ParmTreewidth  = "treewidth"
	ParmNumLatent  = "num_latent"
	ParmMaxParents = "max_parents"
)

// Learner defines a structure learner algorithm
type Learner interface {
	Search() Solution
	SetDefaultParameters()
	SetDataset(*data.Dataset)
	SetFileParameters(parms map[string]string)
	ValidateParameters()
	PrintParameters()
	Treewidth() int
}

// Solution defines a solution interface
type Solution interface {
	Better(interface{}) bool
}

// Create creates a structure learner algorithm
func Create(learnerAlg string) (lr Learner) {
	creators := map[string]func() Learner{}
	if create, ok := creators[learnerAlg]; ok {
		lr = create()
		lr.SetDefaultParameters()
		return
	}
	panic(fmt.Errorf("invalid algorithm option: '%v'", learnerAlg))
}

// Search applies the strategy to find the best solution
func Search(alg Learner, numSolutions, timeAvailable int) Solution {
	var best, current Solution
	if numSolutions <= 0 && timeAvailable <= 0 {
		numSolutions = 1
	}
	if timeAvailable > 0 {
		// TODO: the documentation recommends using NewTimer(d).C instead of time.After
		i := 0
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

			if best == nil || current.Better(best) {
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
		for i := 0; i < numSolutions; i++ {
			current := alg.Search()
			if best == nil || current.Better(best) {
				best = current
			}
		}
	}
	return best
}
