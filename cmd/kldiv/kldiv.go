package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/britojr/lkbn/inference"
	"github.com/britojr/lkbn/model"
	"gonum.org/v1/gonum/floats"
)

func main() {
	var orgFile, compFile string

	if len(os.Args) < 3 {
		fmt.Printf("\n error: missing network files\n\n")
		fmt.Printf("Usage:\n\n")
		fmt.Printf("\tkldiv <original-network> <comparision-network>\n\n")
		os.Exit(1)
	}
	orgFile, compFile = os.Args[1], os.Args[2]
	log.Printf("========== COMPARING NETWORKS =====================\n")
	log.Printf("Comparing %v || %v\n", orgFile, compFile)

	start := time.Now()
	orgNet := model.ReadBNetXML(orgFile)
	compNet := model.ReadCTree(compFile)
	kld := KLDiv(orgNet, compNet)
	elapsed := time.Since(start)

	log.Printf("Time: %v\n", elapsed)
	log.Printf("KL-divergence: %.6f\n", kld)
	log.Printf("---------------------------------------------------\n")
}

// KLDiv computes kl-divergence
func KLDiv(orgNet *model.BNet, compNet *model.CTree) (kld float64) {
	infalg := inference.NewCTreeCalibration(compNet)
	for _, v := range orgNet.Variables() {
		pcond := orgNet.Node(v).Potential().Copy()
		family := pcond.Variables()
		qjoint := infalg.Posterior(family, nil)
		pjoint := orgNet.MarginalizedFamily(v)

		kld += floats.Sum(pjoint.Times(pcond.Log().Minus(qjoint.Log())).Values())
	}
	return kld
}
