package inference

import (
	"testing"

	"github.com/britojr/lkbn/model"
	"github.com/britojr/utl/floats"
)

const tol = 1e-14

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
