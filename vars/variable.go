package vars

import (
	"fmt"
	"strconv"
)

const (
	// DefaultNState default number of states for a variable
	DefaultNState = 2
)

// Var defines variable type
type Var struct {
	id, nstate int
	name       string
	latent     bool
	states     []string
}

// New creates a new variable
func New(id, nstate int, name string, latent bool) (v *Var) {
	v = new(Var)
	v.id = id
	v.name = name
	v.latent = latent
	if name == "" {
		if v.latent {
			v.name = fmt.Sprintf("H%v", v.id)
		} else {
			v.name = fmt.Sprintf("X%v", id)
		}
	}
	v.SetNState(nstate)
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

// SetNState set variable num states
func (v *Var) SetNState(nstate int) {
	v.nstate = nstate
	v.states = make([]string, nstate)
	for i := range v.states {
		v.states[i] = strconv.Itoa(i)
	}
}

// SetStates set variable states names
func (v *Var) SetStates(states []string) {
	v.nstate = len(states)
	v.states = append([]string(nil), states...)
}

// Name return variable's name
func (v Var) Name() string {
	return v.name
}

// SetName set variable name
func (v *Var) SetName(name string) {
	v.name = name
}

// Latent return true for latent variable
func (v Var) Latent() bool {
	return v.latent
}

// SetLatent set latent variable
func (v *Var) SetLatent(latent bool) {
	v.latent = latent
}

func (v Var) String() string {
	return fmt.Sprintf("%v[%v]", v.name, v.nstate)
}

func (v Var) States() (s []string) {
	return v.states
}

func (v Var) StateID(name string) int {
	for i, n := range v.states {
		if n == name {
			return i
		}
	}
	return -1
}
