package lab

import (
	"fmt"
	"sort"
)

func ValidateDenominator(denominator Contract, ir IR) error {
	if denominator.Schema != "gooo/semantic-mutation-lab/denominator/v1" || denominator.CellCount != FixedCells || !denominator.Fixed || len(denominator.Cells) != FixedCells {
		return fmt.Errorf("denominator must contain exactly 12 cells")
	}
	if len(ir.Mutations) != FixedCells {
		return fmt.Errorf("IR must contain exactly 12 mutations")
	}
	mutationsByCell := make(map[string]MutationDecl, len(ir.Mutations))
	for _, mutation := range ir.Mutations {
		mutationsByCell[mutation.Cell] = mutation
	}
	if len(mutationsByCell) != FixedCells {
		return fmt.Errorf("mutation families must bind twelve distinct cells")
	}
	seen := map[string]bool{}
	for _, cell := range denominator.Cells {
		if cell.ID == "" || seen[cell.ID] {
			return fmt.Errorf("denominator has duplicate or empty cell identity")
		}
		seen[cell.ID] = true
		mutation, ok := mutationsByCell[cell.ID]
		if !ok || mutation.ChangedSemanticEdge == "" || mutation.ExpectedDetector == "" {
			return fmt.Errorf("denominator cell %q does not bind the declared mutation", cell.ID)
		}
	}
	ids := make([]string, 0, len(denominator.Cells))
	for _, cell := range denominator.Cells {
		ids = append(ids, cell.ID)
	}
	sort.Strings(ids)
	mutationIDs := make([]string, 0, len(mutationsByCell))
	for id := range mutationsByCell {
		mutationIDs = append(mutationIDs, id)
	}
	sort.Strings(mutationIDs)
	for index := range ids {
		if ids[index] != mutationIDs[index] {
			return fmt.Errorf("denominator does not cover the fixed operator set")
		}
	}
	return nil
}
