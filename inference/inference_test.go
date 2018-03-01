package inference

import (
	"testing"

	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/floats"
	"github.com/britojr/utl/ioutl"
	gfloats "gonum.org/v1/gonum/floats"
)

const tol = 1e-6

func TestRunCalibration(t *testing.T) {
	cases := []struct {
		ctin     string
		evid     map[int]int
		probEvid float64
	}{{
		ctin: `
variables:
  - name : "A"
    card : 2
  - name : "B"
    card : 2
  - name : "C"
    card : 2
  - name : "D"
    card : 2
  - name : "E"
    card : 2
nodes:
  - clqvars : ["A", "B", "C"]
    values : [0.103303, 0.138931, 0.156297, 0.122150, 0.073276, 0.130033, 0.089912,  0.186099]

  - clqvars : ["A","B", "D"]
    values : [0.5646895150969723, 0.5586341386332049, 0.5467096882085889, 0.5722996547239059, 0.43531048490302765, 0.4413658613667951, 0.4532903117914111, 0.4277003452760941]
    parent : ["A", "B", "C"]

  - clqvars : ["B","C", "E"]
    values : [0.5646895150969723, 0.5586341386332049, 0.5467096882085889, 0.5722996547239059, 0.43531048490302765, 0.4413658613667951, 0.4532903117914111, 0.4277003452760941]
    parent : ["A", "B", "C"]
`,
		evid:     map[int]int{0: 1, 1: 1, 2: 1, 3: 1, 4: 1},
		probEvid: (0.186099 * 0.4277003452760941 * 0.4277003452760941),
	}}
	for _, tt := range cases {
		ctin := model.CTreeFromString(tt.ctin)
		inf := NewCTreeCalibration(ctin)
		probEvid := inf.Run(tt.evid)
		if !floats.AlmostEqual(tt.probEvid, probEvid, tol) {
			t.Errorf("wrong prob of evidence %v != %v", tt.probEvid, probEvid)
		}
	}
}

func TestPosterior(t *testing.T) {
	ctin1 := `
variables:
  - name : "A"
    card : 2
  - name : "B"
    card : 2
  - name : "C"
    card : 2
  - name : "D"
    card : 2
  - name : "E"
    card : 2
nodes:
  - clqvars : ["A", "B", "C"]
    values : [0.103303, 0.138931, 0.156297, 0.122150, 0.073276, 0.130033, 0.089912,  0.186099]

  - clqvars : ["A","B", "D"]
    values : [0.5646895150969723, 0.5586341386332049, 0.5467096882085889, 0.5722996547239059, 0.43531048490302765, 0.4413658613667951, 0.4532903117914111, 0.4277003452760941]
    parent : ["A", "B", "C"]

  - clqvars : ["B","C", "E"]
    values : [0.5646895150969723, 0.5586341386332049, 0.5467096882085889, 0.5722996547239059, 0.43531048490302765, 0.4413658613667951, 0.4532903117914111, 0.4277003452760941]
    parent : ["A", "B", "C"]
`
	cases := []struct {
		ctin     string
		evid     map[int]int
		vs       vars.VarList
		postVals []float64
	}{{
		ctin:     ctin1,
		evid:     map[int]int{0: 1, 1: 1, 2: 1, 3: 1, 4: 1},
		vs:       []*vars.Var{},
		postVals: []float64{(0.186099 * 0.4277003452760941 * 0.4277003452760941)},
	}, {
		ctin:     ctin1,
		evid:     map[int]int{},
		vs:       []*vars.Var{},
		postVals: []float64{1.0},
	}, {
		ctin:     ctin1,
		evid:     map[int]int{},
		vs:       vars.NewList([]int{0, 4}, []int{2, 2}),
		postVals: []float64{0.23716426661272494, 0.32428473338727504, 0.18562373338727503, 0.25292826661272494},
	}}
	for _, tt := range cases {
		ctin := model.CTreeFromString(tt.ctin)
		inf := NewCTreeCalibration(ctin)
		posterior := inf.Posterior(tt.vs, tt.evid)
		if len(posterior.Values()) != len(tt.postVals) {
			t.Errorf("wrong posterior values:\n%v\n!=\n%v\n", tt.postVals, posterior.Values())
		} else {
			for i, v := range tt.postVals {
				if !floats.AlmostEqual(v, posterior.Values()[i], tol) {
					t.Errorf("wrong posterior values [%v] %v != %v", i, v, posterior.Values()[i])
				}
			}
		}
	}
}

var ctreeN5K3 = `
variables:
- {name: node0,  card: 2}
- {name: node1,  card: 2}
- {name: node2,  card: 2}
- {name: node3,  card: 2}
- {name: node4,  card: 2}
nodes:
- clqvars: [node0, node1, node3, node4]
  values: [1.1157054432479998E-02, 8.882308761788003E-03, 2.3162421377399996E-03, 6.144352496990398E-02, 6.119903275199999E-04, 1.6292250287821203E-01, 5.310131022599999E-04, 1.5708063390096E-02, 2.56896382210776E-01, 3.8671336423776E-02, 2.3493122640931997E-02, 2.8890730545011997E-02, 2.0398697302922397E-01, 7.453515193622401E-02, 8.800722211906799E-02, 2.1946381094987994E-02]
  parent: []
- clqvars: [node0, node1, node2]
  values: [3.264E-01, 3.579E-01, 5.806E-01, 5.468E-01, 6.736E-01, 6.421E-01, 4.194E-01, 4.532E-01]
  parent: [node0, node1, node3, node4]
`

func TestPosterior2(t *testing.T) {
	ctreeFile := ioutl.TempFile("inference_test", ctreeN5K3)
	ct := model.ReadCTree(ctreeFile)
	pjoint := ct.Nodes()[0].Potential().Copy()
	for _, nd := range ct.Nodes()[1:] {
		pjoint.Times(nd.Potential())
	}
	inf := NewCTreeCalibration(ct)
	// compare joint
	got := inf.Posterior(ct.Variables(), nil)
	if !gfloats.EqualApprox(pjoint.Values(), got.Values(), tol) {
		for i := range pjoint.Values() {
			t.Errorf("wrong result [%v]: %v\n", i, pjoint.Values()[i]-got.Values()[i])
		}
	}
	// compare marginals
	for _, v := range ct.Variables() {
		want := pjoint.Copy().Marginalize(v)
		got = inf.Posterior(vars.VarList{v}, nil)
		if !gfloats.EqualApprox(want.Values(), got.Values(), tol) {
			for i := range want.Values() {
				t.Errorf("wrong result [%v]: %v\n", i, want.Values()[i]-got.Values()[i])
			}
		}
	}
}
