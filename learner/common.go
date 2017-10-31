package learner

import (
	"log"

	"github.com/britojr/lkbn/data"
	"github.com/britojr/utl/conv"
)

// common defines default behaviours
type common struct {
	tw int           // treewidth
	nv int           // number of variables
	nl int           // number of latent variables
	ds *data.Dataset // dataset
}

func newCommon() *common {
	s := new(common)
	return s
}

func (s *common) SetDataset(ds *data.Dataset) {
	s.ds = ds
}

func (s *common) SetDefaultParameters() {
	s.tw = 3
}

func (s *common) SetFileParameters(parms map[string]string) {
	if tw, ok := parms[ParmTreewidth]; ok {
		s.tw = conv.Atoi(tw)
	}
	if nl, ok := parms[ParmNumLatent]; ok {
		s.nl = conv.Atoi(nl)
	}
}

func (s *common) ValidateParameters() {
	if s.tw <= 0 || s.nv < s.tw+2 {
		log.Printf("n=%v, tw=%v\n", s.nv, s.tw)
		log.Panic("Invalid treewidth! Choose values such that: n >= tw+2 and tw > 0")
	}
	if s.nl < 0 {
		log.Printf("n=%v, tw=%v\n", s.nv, s.tw)
		log.Panicf("Invalid number of latent variables: '%v'", s.nl)
	}
}

func (s *common) PrintParameters() {
	log.Printf(" ========== ALGORITHM PARAMETERS ================ \n")
	log.Printf("number of variables: %v\n", s.nv)
	log.Printf("%v: %v\n", ParmTreewidth, s.tw)
	log.Printf("%v: %v\n", ParmNumLatent, s.nl)
}

func (s *common) Treewidth() int {
	return s.tw
}
