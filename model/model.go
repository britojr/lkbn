package model

// Model defines a probabilistic model
type Model interface {
	ToCTree() *CTree
}
