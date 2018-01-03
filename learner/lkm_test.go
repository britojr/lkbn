package learner

import (
	"testing"

	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
	"github.com/britojr/utl/ioutl"
)

func TestCreateLKM1LStruct(t *testing.T) {
	vs := vars.NewList([]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil)
	strct := `
variables:
- {name: X0,  card: 2}
- {name: X1,  card: 2}
- {name: X2,  card: 2}
- {name: X3,  card: 2}
- {name: X4,  card: 2}
- {name: X5,  card: 2}
- {name: X6,  card: 2}
- {name: X7,  card: 2}
- {name: X8,  card: 2}
- {name: X9,  card: 2}
- {name: X10, card: 2}
- {name: Y11,  card: 3, latent: true}
nodes:
- clqvars: [X1, X5, X7, Y11]
- clqvars: [X2, X3, Y11]
  parent: [X1, X5, X7, Y11]
- clqvars: [X4, X6, X8, Y11]
  parent: [X1, X5, X7, Y11]
- clqvars: [X0, X9, X10, Y11]
  parent: [X1, X5, X7, Y11]
`
	cases := []struct {
		gs []vars.VarList
		lv *vars.Var
		ct *model.CTree
	}{{
		[]vars.VarList{
			[]*vars.Var{vs[1], vs[5], vs[7]},
			[]*vars.Var{vs[2], vs[3]},
			[]*vars.Var{vs[4], vs[6], vs[8]},
			[]*vars.Var{vs[0], vs[9], vs[10]},
		},
		vars.New(11, 3, "Y11", true),
		model.CTreeFromString(strct),
	}}
	for _, tt := range cases {
		ct := createLKM1LStruct(tt.gs, tt.lv)
		if !tt.ct.Equal(ct) {
			t.Errorf("different tree:\n%v\n-----\n%v\n", tt.ct, ct)
		}
	}
}

type fakeLearner struct {
	maxcard int
}

func (f fakeLearner) SetProperties(props map[string]string) {}
func (f fakeLearner) PrintProperties()                      {}
func (f fakeLearner) Run(m *model.CTree, evset []map[int]int) (*model.CTree, float64, int) {
	vs := m.Variables()
	c := 1
	for _, v := range vs {
		if v.Latent() {
			c = v.NState()
			break
		}
	}
	if c > f.maxcard {
		c = f.maxcard
	}
	m.SetScore(-1000 / float64(c))
	return m, m.Score(), 1
}

func TestLearnLKM1L(t *testing.T) {
	content := `A,B,C,D,E,F,G,H,I,J,K
0,0,0,0,0,0,0,0,0,0,0
1,0,1,1,1,1,1,0,1,1,1
1,1,1,1,1,1,1,1,1,1,1
0,0,0,0,0,0,0,0,0,0,0
0,0,0,0,0,0,0,0,0,0,0
0,0,0,0,1,0,0,0,0,0,0
0,0,0,0,0,0,0,0,0,0,0
0,0,0,0,0,0,0,0,0,0,0
0,0,0,0,0,0,0,0,0,0,0
`
	vs := vars.NewList([]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil)
	fname := ioutl.TempFile("lkm_test", content)
	cases := []struct {
		gs      []vars.VarList
		ds      *data.Dataset
		maxcard int
	}{{
		[]vars.VarList{
			[]*vars.Var{vs[1], vs[5], vs[7]},
			[]*vars.Var{vs[2], vs[3]},
			[]*vars.Var{vs[4], vs[6], vs[8]},
			[]*vars.Var{vs[0], vs[9], vs[10]},
		},
		data.NewDataset(fname),
		5,
	}}
	for _, tt := range cases {
		ct := learnLKM1L(tt.gs, tt.ds, fakeLearner{tt.maxcard})
		vs := ct.Variables()
		if len(vs) != len(tt.ds.Variables())+1 {
			t.Errorf("latent variable wasn't created: %v", vs)
		}
		lv := ct.Variables()[len(tt.ds.Variables())]
		if lv.NState() != tt.maxcard {
			t.Errorf("stoped at wrong cardinality: %v != %v", tt.maxcard, lv.NState())
		}
	}
}
