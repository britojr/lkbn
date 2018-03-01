package main

import (
	"fmt"
	"log"
	"os"

	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/scores"
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

	orgNet := model.ReadBNetXML(orgFile)
	compNet := model.ReadCTree(compFile)

	log.Printf("KL-Divergence: %.6f\n", scores.KLDiv(orgNet, compNet))
	log.Printf("KLDiv brute force: %.6f\n", scores.KLDivBruteForce(orgNet, compNet))
	log.Printf("---------------------------------------------------\n")
}
