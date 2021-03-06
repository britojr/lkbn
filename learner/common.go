package learner

import (
	"log"
	"strings"

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
	s.nv = len(ds.Variables())
	s.vs = ds.Variables().Copy()
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
	} else {
		if lvStr, ok := props[ParmLatentVars]; ok {
			latentVars := conv.Satoi(strings.FieldsFunc(lvStr, func(r rune) bool {
				return r == ',' || r == ' '
			}))
			s.nl = len(latentVars)
			s.addLatVars(latentVars...)
		}
	}
	s.paramLearner.SetProperties(props)
	s.paramLearner.PrintProperties()
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

func (s *common) addLatVars(latentVars ...int) {
	if len(latentVars) == 0 {
		for i := 0; i < s.nl; i++ {
			v := vars.New(len(s.vs), s.lstates, "", true)
			s.vs.Add(v)
		}
	} else {
		for _, card := range latentVars {
			v := vars.New(len(s.vs), card, "", true)
			s.vs.Add(v)
		}
	}
}
