package learner

import (
	"log"

	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/emlearner"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/conv"
)

// common defines default behaviours
type common struct {
	tw           int                 // treewidth
	nv           int                 // number of observed variables
	nl           int                 // number of latent variables
	lstates      int                 // default number of states of latent variables
	ds           *data.Dataset       // dataset
	vs           vars.VarList        // variables
	props        map[string]string   // property map
	paramLearner emlearner.EMLearner // parameter learner
}

func newCommon() *common {
	s := new(common)
	s.paramLearner = emlearner.New()
	return s
}

func (s *common) SetDataset(ds *data.Dataset) {
	s.ds = ds
	s.vs = make([]*vars.Var, len(ds.Variables()))
	s.nv = len(ds.Variables())
	copy(s.vs, ds.Variables())
}

func (s *common) SetDefaultParameters() {
	s.tw = 3
	s.lstates = 2
}

func (s *common) SetFileParameters(props map[string]string) {
	if tw, ok := props[ParmTreewidth]; ok {
		s.tw = conv.Atoi(tw)
	}
	if nl, ok := props[ParmNumLatent]; ok {
		s.nl = conv.Atoi(nl)
		s.addLatVars()
	}
	s.paramLearner.SetProperties(props)
	s.props = props
}

func (s *common) ValidateParameters() {
	if s.tw <= 0 || len(s.vs) < s.tw+2 {
		log.Printf("n=%v, tw=%v\n", len(s.vs), s.tw)
		log.Panic("Invalid treewidth! Choose values such that: n >= tw+2 and tw > 0")
	}
	if s.nl < 0 {
		log.Panicf("Invalid number of latent variables: '%v'", s.nl)
	}
}

func (s *common) PrintParameters() {
	log.Printf("============ ALGORITHM PARAMETERS ================\n")
	log.Printf("total number of variables: %v\n", len(s.vs))
	log.Printf("%v: %v\n", ParmTreewidth, s.tw)
	log.Printf("%v: %v\n", ParmNumLatent, s.nl)
}

func (s *common) Treewidth() int {
	return s.tw
}

func (s *common) addLatVars() {
	for i := 0; i < s.nl; i++ {
		v := vars.New(len(s.vs), s.lstates)
		v.SetLatent(true)
		s.vs = append(s.vs, v)
	}
}
