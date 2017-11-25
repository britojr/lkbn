package vars

import "fmt"

const (
	// DefaultNState default number of states for a variable
	DefaultNState = 2
)

// Var defines variable type
type Var struct {
	id, nstate int
	name       string
	latent     bool
}

// New creates a new variable
func New(id, nstate int) (v *Var) {
	v = new(Var)
	v.id = id
	v.nstate = nstate
	v.name = fmt.Sprintf("x%v", id)
	return
}

// ID variable's id
func (v Var) ID() int {
	return v.id
}

// NState variable's number of states
func (v Var) NState() int {
	return v.nstate
}

// Name return variable's name
func (v Var) Name() string {
	return v.name
}

// SetName set variable name
func (v Var) SetName(name string) {
	v.name = name
}

// Latent return true for latent variable
func (v Var) Latent() bool {
	return v.latent
}

// SetLatent set latent variable
func (v Var) SetLatent(latent bool) {
	v.latent = latent
}

func (v Var) String() string {
	return fmt.Sprintf("x%v[%v]", v.id, v.nstate)
}
