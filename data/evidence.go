package data

// Evidence ..
type Evidence map[int]int

// EvidenceSet ...
type EvidenceSet interface {
	Observations() []Evidence
}
