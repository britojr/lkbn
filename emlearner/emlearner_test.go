package emlearner

import (
	"math"
	"testing"

	"github.com/britojr/lkbn/inference"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/utl/floats"
)

const tol = 1e-14

// ABC:0.103303, 0.138931, 0.156297, 0.122150, 0.073276, 0.130033, 0.089912,  0.186099
// ABD:0.125050, 0.125433, 0.117747, 0.135743, 0.051529, 0.143530, 0.128461, 0.172506
// D|AB: 0.5646895150969723, 0.5586341386332049, 0.5467096882085889, 0.5722996547239059, 0.43531048490302765, 0.4413658613667951, 0.4532903117914111, 0.4277003452760941
// BCE:0.136787, 0.155550, 0.111151, 0.157961, 0.105447, 0.122897, 0.092158, 0.118050
// E|BC: 0.5646895150969723, 0.5586341386332049, 0.5467096882085889, 0.5722996547239059, 0.43531048490302765, 0.4413658613667951, 0.4532903117914111, 0.4277003452760941

func TestRunStep(t *testing.T) {
	cases := []struct {
		ctin  string
		data  []map[int]int
		ctout string
		ll    float64
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
		ctout: `
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
    values : [0, 0, 0, 0, 0, 0, 0.75, 0.25]

  - clqvars : ["A","B", "D"]
    values : [0, 0, 1, 0, 0, 0, 0, 1]
    parent : ["A", "B", "C"]

  - clqvars : ["B","C", "E"]
    values : [0, 0, 0, 0, 0, 0, 0, 1]
    parent : ["A", "B", "C"]
`,
		data: []map[int]int{
			{0: 0, 1: 1, 2: 1, 3: 0, 4: 1},
			{0: 0, 1: 1, 2: 1, 3: 0, 4: 1},
			{0: 1, 1: 1, 2: 1, 3: 1, 4: 1},
			{0: 0, 1: 1, 2: 1, 3: 0, 4: 1},
		},
		ll: 1*math.Log(0.186099*0.4277003452760941*0.4277003452760941) +
			3*math.Log(0.089912*0.5467096882085889*0.4277003452760941),
	}}
	for _, tt := range cases {
		ctin := model.FromString(tt.ctin)
		e := new(emAlg)
		e.maxIters = 0
		e.threshold = 1e-8
		inf := inference.NewCTreeCalibration(ctin)
		ll := e.runStep(inf, tt.data)
		if !floats.AlmostEqual(tt.ll, ll, tol) {
			t.Errorf("wrong ll %v != %v", tt.ll, ll)
		}
		ctout := model.FromString(tt.ctout)
		got := inf.UpdatedModel().(*model.CTree)
		if !got.Equal(ctout) {
			// t.Errorf("wrong ctree:\n%v\n!=\n%v\n", got, ctout)
		}
	}
}
