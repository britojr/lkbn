package main

import (
	"fmt"
	"log"
	"os"
	"time"

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

	start := time.Now()
	orgNet := model.ReadBNetXML(orgFile)
	compNet := model.ReadCTree(compFile)
	kld := scores.KLDiv(orgNet, compNet)
	elapsed := time.Since(start)

	log.Printf("Time: %v\n", elapsed)
	log.Printf("KL-divergence: %.6f\n", kld)
	log.Printf("---------------------------------------------------\n")
}
