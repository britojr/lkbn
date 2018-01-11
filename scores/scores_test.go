package scores

import (
	"testing"

	"github.com/britojr/lkbn/model"
)

func TestComputeModelSize(t *testing.T) {
	struct1 := `
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
- {name: Y12,  card: 5, latent: true}
nodes:
- clqvars: [Y11, Y12]
- clqvars: [X1, X5, X7, Y12]
  parent: [Y11, Y12]
- clqvars: [X2, X3, Y12]
  parent: [Y11, Y12]
- clqvars: [X4, X6, X8, Y11]
  parent: [Y11, Y12]
- clqvars: [X0, X9, X10, Y12]
  parent: [Y11, Y12]
`
	cases := []struct {
		ct   *model.CTree
		want int
	}{{
		model.CTreeFromString(struct1),
		120,
	}}
	for _, tt := range cases {
		got := computeModelSize(tt.ct)
		if got != tt.want {
			t.Errorf("wrong model size: %v != %v", tt.want, got)
		}
	}
}
