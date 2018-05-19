package main

import (
	"fmt"
	"log"
	"os"

	"github.com/britojr/lkbn/data"
	"github.com/britojr/lkbn/model"
	"github.com/britojr/lkbn/scores"
)

func main() {
	var orgFile, compFile, dsFile string

	if len(os.Args) < 3 {
		fmt.Printf("\n error: missing network files\n\n")
		fmt.Printf("Usage:\n\n")
		fmt.Printf("\tkldiv <original-network> <comparision-network>\n\n")
		os.Exit(1)
	}
	orgFile, compFile = os.Args[1], os.Args[2]
	if len(os.Args) > 3 {
		dsFile = os.Args[3]
	}
	log.Printf("========== COMPARING NETWORKS =====================\n")
	log.Printf("Comparing %v || %v\n", orgFile, compFile)

	orgNet := model.ReadBNetXML(orgFile)
	compNet := model.ReadCTreeXML(compFile)

	log.Printf("KL-Divergence: %E\n", scores.KLDiv(orgNet, compNet))
	log.Printf("KLDiv brute force: %E\n", scores.KLDivBruteForce(orgNet, compNet))
	if len(dsFile) > 0 {
		ds := data.NewDataset(dsFile, "")
		log.Printf("KLDiv empirical: %E\n", scores.KLDivEmpirical(orgNet, compNet, ds.IntMaps()))
		log.Printf("KLDiv empirical no inf: %E\n", scores.KLDivEmpNoInf(orgNet, compNet, ds.IntMaps()))
	}
	log.Printf("---------------------------------------------------\n")
}
