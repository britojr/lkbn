package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	var bnFile, ctFile string

	if len(os.Args) < 3 {
		fmt.Printf("\n error: missing network files\n\n")
		fmt.Printf("Usage:\n\n")
		fmt.Printf("\tkldiv <original-network> <comparision-network>\n\n")
		os.Exit(1)
	}
	bnFile, ctFile = os.Args[1], os.Args[2]
	log.Printf("========== COMPARING NETWORKS =====================\n")
	log.Printf("Comparing %v || %v\n", bnFile, ctFile)

	start := time.Now()
	kld := 0.0
	elapsed := time.Since(start)

	log.Printf("Time: %v\n", elapsed)
	log.Printf("KL-divergence: %.6f\n", kld)
	log.Printf("---------------------------------------------------\n")
}
