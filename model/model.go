package model

// Model defines a probabilistic model
type Model interface {
	ToCTree() *CTree
	SetScore(float64)
	Score() float64
}
