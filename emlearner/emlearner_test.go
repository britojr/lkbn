package emlearner

import (
	"testing"

	"github.com/britojr/lkbn/inference"
	"github.com/britojr/lkbn/model"
)

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
  - clqvars : ["A"]
    values : [0.999, 0.001]
    parent : ["A", "B", "C"]
  - clqvars : ["B"]
    values : [0.998, 0.002]
    parent : ["A", "B", "C"]
  - clqvars : ["A", "B", "C"]
    values : [0.999, 0.06, 0.71, 0.05, 0.001, 0.94, 0.29, 0.95]
  - clqvars : ["C", "D"]
    values : [0.95, 0.10, 0.05, 0.90]
    parent : ["A", "B", "C"]
  - clqvars : ["C", "E"]
    values : [0.99, 0.30, 0.01, 0.70]
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
  - clqvars : ["A"]
    values : [0.75, 0.25]
    parent : ["A", "B", "C"]
  - clqvars : ["B"]
    values : [0, 1]
    parent : ["A", "B", "C"]
  - clqvars : ["A", "B", "C"]
    values : [0, 0, 0, 0, 0, 0, 0.75, 0.25]
  - clqvars : ["C", "D"]
    values : [0, 0.75, 0, 0.25]
    parent : ["A", "B", "C"]
  - clqvars : ["C", "E"]
    values : [0, 0, 0, 1]
    parent : ["A", "B", "C"]
`,
		data: []map[int]int{
			{0: 0, 1: 1, 2: 1, 3: 0, 4: 1},
			{0: 0, 1: 1, 2: 1, 3: 0, 4: 1},
			{0: 1, 1: 1, 2: 1, 3: 1, 4: 1},
			{0: 0, 1: 1, 2: 1, 3: 0, 4: 1},
		},
		ll: -4.498681156950466,
	}}
	for _, tt := range cases {
		ctin := model.FromString(tt.ctin)
		e := new(emAlg)
		e.maxIters = 0
		e.threshold = 1e-8
		inf := inference.NewCTreeCalibration(ctin)
		ll := e.runStep(inf, tt.data)
		if tt.ll != ll {
			t.Errorf("wrong ll %v != %v", tt.ll, ll)
		}
		ctout := model.FromString(tt.ctout)
		got := inf.UpdatedModel().(*model.CTree)
		if !got.Equal(ctout) {
			t.Errorf("wrong ctree:\n%v\n!=\n%v\n", got, ctout)
		}
	}
}
