package factor

import (
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/britojr/lkbn/vars"
	"gonum.org/v1/gonum/floats"
)

var (
	// ErrZeroSum error used to inform that noramlization tried to divede by zero
	ErrZeroSum = errors.New("Normalization group add up to zero")
)

const (
	eqtol = 1e-14
)

// Factor defines a function that maps a joint of categorical variables to float values
// every factor operation, if not explicitly stated otherwise, modifies the given factor
type Factor struct {
	values []float64
	vs     vars.VarList
}

var (
	opAdd = func(a, b float64) float64 { return a + b }
	opSub = func(a, b float64) float64 { return a - b }
	opDiv = func(a, b float64) float64 {
		if b == 0 {
			panic("factor: op division by zero")
		}
		return a / b
	}
	opMul = func(a, b float64) float64 { return a * b }
)

var seed = func() int64 {
	return time.Now().UnixNano()
}

// New creates a new factor with uniform distribution for given variables
// a factor with no variables have a value of one
func New(xs ...*vars.Var) (f *Factor) {
	f = new(Factor)
	f.vs = vars.VarList(xs).Copy()
	f.values = make([]float64, f.vs.NStates())
	tot := float64(len(f.values))
	for i := range f.values {
		f.values[i] = 1 / tot
	}
	return
}

// NewZeroes creates a new factor and leaves values set to zeroes
func NewZeroes(xs ...*vars.Var) (f *Factor) {
	f = new(Factor)
	f.vs = vars.VarList(xs).Copy()
	f.values = make([]float64, f.vs.NStates())
	return
}

// NewIndicator creates a new factor with just one state value set to one
func NewIndicator(xs *vars.Var, state int) (f *Factor) {
	f = new(Factor)
	f.vs = []*vars.Var{xs}
	f.values = make([]float64, f.vs.NStates())
	f.values[state] = 1
	return
}

// Copy returns a copy of f
func (f *Factor) Copy() (g *Factor) {
	g = new(Factor)
	g.vs = f.vs.Copy()
	g.values = make([]float64, len(f.values))
	copy(g.values, f.values)
	return
}

// SetValues sets the given values to the factor
func (f *Factor) SetValues(values []float64) *Factor {
	copy(f.values, values)
	return f
}

// Values return reference for factor values
func (f *Factor) Values() []float64 {
	return f.values
}

// Variables return reference for factor variables
func (f *Factor) Variables() vars.VarList {
	return f.vs
}

// RandomDistribute sets values with a random distribution
func (f *Factor) RandomDistribute(xs ...*vars.Var) *Factor {
	r := rand.New(rand.NewSource(seed()))
	for i := range f.values {
		f.values[i] = r.Float64()
		for f.values[i] <= 0 {
			f.values[i] = r.Float64()
		}
	}
	f.Normalize(xs...)
	return f
}

// UniformDistribute sets values with a uniform distribution
func (f *Factor) UniformDistribute(xs ...*vars.Var) *Factor {
	tot := float64(len(f.values))
	for i := range f.values {
		f.values[i] = 1 / tot
	}
	f.Normalize(xs...)
	return f
}

// ResetValues realocates values slice
func (f *Factor) ResetValues() *Factor {
	f.values = make([]float64, f.vs.NStates())
	return f
}

// Log applies log on factor values
func (f *Factor) Log() *Factor {
	for i, v := range f.values {
		f.values[i] = math.Log(v)
	}
	return f
}

// Plus adds g to f
func (f *Factor) Plus(g *Factor) *Factor {
	return f.operationIn(g, opAdd)
}

// Minus subtracts g from f (returns f = f - g)
func (f *Factor) Minus(g *Factor) *Factor {
	return f.operationIn(g, opSub)
}

// Times multiplies f by g
func (f *Factor) Times(g *Factor) *Factor {
	return f.operationIn(g, opMul)
}

// operation applies given operation as f = f op g
func (f *Factor) operationIn(g *Factor, op func(a, b float64) float64) *Factor {
	if f.vs.Equal(g.vs) {
		return f.operationEq(g, op)
	}
	return f.operationTr(f.Copy(), g, op)
}

// operationEq applies given operation as f = f op g assuming equal sets of variables
func (f *Factor) operationEq(g *Factor, op func(a, b float64) float64) *Factor {
	for i, v := range g.values {
		f.values[i] = op(f.values[i], v)
	}
	return f
}

// operationTr returns applies given operation as operation f = f1 op f2
func (f *Factor) operationTr(f1, f2 *Factor, op func(a, b float64) float64) *Factor {
	f.vs = f1.vs.Union(f2.vs)
	f.values = make([]float64, f.vs.NStates())
	ixf1 := vars.NewIndexFor(f1.vs, f.vs)
	ixf2 := vars.NewIndexFor(f2.vs, f.vs)
	for i := range f.values {
		f.values[i] = op(f1.values[ixf1.I()], f2.values[ixf2.I()])
		ixf1.Next()
		ixf2.Next()
	}
	return f
}

// Normalize normalizes f so the values add up to one
// the normalization can be relative to a list of variables
//    ex: P(X,Y,Z).Normalize(X) => P(X|Y,Z)
func (f *Factor) Normalize(xs ...*vars.Var) (*Factor, error) {
	var err error
	if len(xs) == 0 {
		return f.normalizeAll()
	}
	condVars := f.vs.Diff(xs)
	if len(condVars) == 0 {
		return f.normalizeAll()
	}
	ixf := vars.NewIndexFor(condVars, f.vs)
	sums := make([]float64, condVars.NStates())
	for _, v := range f.values {
		sums[ixf.I()] += v
		ixf.Next()
	}
	ixf.Reset()
	for i := range f.values {
		if sums[ixf.I()] != 0 {
			f.values[i] /= sums[ixf.I()]
		} else {
			err = ErrZeroSum
			f.values[i] = 0
		}
		ixf.Next()
	}
	return f, err
}

func (f *Factor) normalizeAll() (*Factor, error) {
	var err error
	sum := floats.Sum(f.values)
	if sum != 0 {
		for i := range f.values {
			f.values[i] /= sum
		}
	} else {
		err = ErrZeroSum
	}
	return f, err
}

// SumOut sums out the given variables
func (f *Factor) SumOut(xs ...*vars.Var) *Factor {
	if len(xs) == 0 {
		return f
	}
	oldVs := f.vs
	oldVal := f.values
	f.vs = oldVs.Diff(xs)
	f.values = make([]float64, f.vs.NStates())
	ixf := vars.NewIndexFor(f.vs, oldVs)
	for _, v := range oldVal {
		f.values[ixf.I()] += v
		ixf.Next()
	}
	return f
}

// Marginalize projects the distribution in the given variables
func (f *Factor) Marginalize(xs ...*vars.Var) *Factor {
	return f.SumOut(f.vs.Diff(xs)...)
}

// Reduce silences the values that are not compatible with the given evidence
func (f *Factor) Reduce(e map[int]int) *Factor {
	step, ind := 1, 0
	var rvs []*vars.Var
	for _, v := range f.vs {
		if a, ok := e[v.ID()]; ok {
			ind += a * step
			step *= v.NState()
			rvs = append(rvs, v)
		}
	}
	if len(rvs) == 0 {
		return f
	}
	if len(rvs) == len(f.vs) {
		for i := range f.values {
			if i != ind {
				f.values[i] = 0
			}
		}
		return f
	}
	ixr := vars.NewIndexFor(rvs, f.vs)
	for i := range f.values {
		if ixr.I() != ind {
			f.values[i] = 0
		}
		ixr.Next()
	}
	return f
}

// Get returns the value corresponding to the given attribution
// the attribution must cover the factor scope
func (f *Factor) Get(e map[int]int) float64 {
	step, ind := 1, 0
	for _, v := range f.vs {
		if a, ok := e[v.ID()]; ok {
			ind += a * step
			step *= v.NState()
		} else {
			panic("missing variable of the scope")
		}
	}
	return f.values[ind]
}

func (f *Factor) Equal(g *Factor) bool {
	if !f.vs.Equal(g.vs) {
		return false
	}
	return floats.EqualApprox(f.values, g.values, eqtol)
}
