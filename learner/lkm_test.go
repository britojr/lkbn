package learner

import (
	"testing"

	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/vars"
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
