package lab

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ParseSource(path string) (SourceDecl, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SourceDecl{}, err
	}
	decl := SourceDecl{SourceDigest: DigestBytes(data)}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(strings.SplitN(scanner.Text(), "#", 2)[0])
		line = strings.TrimSpace(strings.SplitN(line, "//", 2)[0])
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		switch fields[0] {
		case "gooo":
			if len(fields) != 3 || fields[1] != "semantic_mutation_lab" || fields[2] != "v1" {
				return SourceDecl{}, fmt.Errorf("line %d: invalid gooo header", lineNumber)
			}
			decl.Schema = "gooo/semantic-mutation-lab/source/v1"
			decl.Version = fields[2]
		case "denominator":
			values, err := parseKeyValues(fields[1:])
			if err != nil {
				return SourceDecl{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			decl.DenominatorID = values["id"]
			decl.CellCount, err = parseInt(values, "cell_count")
			if err != nil {
				return SourceDecl{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
		case "authority":
			values, err := parseKeyValues(fields[1:])
			if err != nil {
				return SourceDecl{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			decl.Authority.RepositoryWrites, err = parseInt(values, "repository_writes")
			if err != nil {
				return SourceDecl{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			decl.Authority.LocalTestExecutions, err = parseInt(values, "local_test_executions")
			if err != nil {
				return SourceDecl{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			decl.Authority.CrossProjectRequiredGates, err = parseInt(values, "cross_project_required_gates")
			if err != nil {
				return SourceDecl{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
		case "precedence":
			if len(fields) != 2 {
				return SourceDecl{}, fmt.Errorf("line %d: invalid precedence", lineNumber)
			}
			decl.Precedence = strings.Split(fields[1], ">")
		case "unknown_fields":
			if len(fields) != 2 {
				return SourceDecl{}, fmt.Errorf("line %d: invalid unknown_fields", lineNumber)
			}
			decl.UnknownFields = strings.Split(fields[1], ",")
		case "cell":
			values, err := parseKeyValues(fields[1:])
			if err != nil {
				return SourceDecl{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			ordinal, err := parseInt(values, "ordinal")
			if err != nil {
				return SourceDecl{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			depends := []string{}
			if values["depends_on"] != "-" && values["depends_on"] != "" {
				depends = strings.Split(values["depends_on"], ",")
			}
			decl.Cells = append(decl.Cells, CellDecl{Ordinal: ordinal, ID: values["id"], SemanticEdge: values["edge"], DependsOn: depends})
		case "mutation":
			values, err := parseKeyValues(fields[1:])
			if err != nil {
				return SourceDecl{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			ordinal, err := parseInt(values, "ordinal")
			if err != nil {
				return SourceDecl{}, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			decl.Mutations = append(decl.Mutations, MutationDecl{Ordinal: ordinal, ID: values["id"], Family: values["family"], Cell: values["cell"], ChangedSemanticEdge: values["edge"], Operator: values["operator"], ExpectedDetector: values["detector"]})
		default:
			return SourceDecl{}, fmt.Errorf("line %d: unknown declaration %q", lineNumber, fields[0])
		}
	}
	if err := scanner.Err(); err != nil {
		return SourceDecl{}, err
	}
	return decl, nil
}

func parseKeyValues(fields []string) (map[string]string, error) {
	values := make(map[string]string, len(fields))
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid key/value %q", field)
		}
		values[parts[0]] = strings.Trim(parts[1], "\"")
	}
	return values, nil
}

func parseInt(values map[string]string, key string) (int, error) {
	value, ok := values[key]
	if !ok {
		return 0, fmt.Errorf("missing %s", key)
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q", key, value)
	}
	return n, nil
}

func LoadContract(path string) (Contract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Contract{}, err
	}
	var contract Contract
	if err := json.Unmarshal(data, &contract); err != nil {
		return Contract{}, fmt.Errorf("decode contract: %w", err)
	}
	if contract.Schema != "gooo/semantic-mutation-lab/denominator/v1" {
		return Contract{}, fmt.Errorf("unexpected contract schema %q", contract.Schema)
	}
	return contract, nil
}

func ContractDigest(contract Contract) (string, error) {
	return DigestValue(contract)
}

func ValidateDeclarations(source SourceDecl, contract Contract) error {
	if source.Schema != "gooo/semantic-mutation-lab/source/v1" || source.Version != "v1" {
		return fmt.Errorf("invalid source declaration")
	}
	if source.DenominatorID != contract.ID || source.CellCount != FixedCells || contract.CellCount != FixedCells || !contract.Fixed {
		return fmt.Errorf("fixed denominator declaration mismatch")
	}
	if len(source.Cells) != FixedCells || len(contract.Cells) != FixedCells || len(source.Mutations) != FixedCells || len(contract.Families) != FixedCells {
		return fmt.Errorf("expected exactly %d cells and mutations", FixedCells)
	}
	if len(source.UnknownFields) != 6 || strings.Join(source.UnknownFields, ",") != "stage,step,reason,unknown_class,next_operation,blocked_by" {
		return fmt.Errorf("UNKNOWN six-field contract mismatch")
	}
	if strings.Join(source.Precedence, ">") != "REFUTED>UNKNOWN>CLOSED" {
		return fmt.Errorf("resolution precedence mismatch")
	}
	seenCells, seenFamilies := map[string]bool{}, map[string]bool{}
	for i := 0; i < FixedCells; i++ {
		cell, expectedCell := source.Cells[i], contract.Cells[i]
		if cell.Ordinal != i+1 || expectedCell.Ordinal != i+1 || cell.ID == "" || seenCells[cell.ID] || cell.ID != expectedCell.ID || cell.SemanticEdge != expectedCell.SemanticEdge || !sameStrings(cell.DependsOn, expectedCell.DependsOn) {
			return fmt.Errorf("cell %d does not match fixed contract", i+1)
		}
		seenCells[cell.ID] = true
		mutation, expectedMutation := source.Mutations[i], contract.Families[i]
		if mutation.Ordinal != i+1 || expectedMutation.Ordinal != i+1 || mutation.ID == "" || mutation.Family == "" || seenFamilies[mutation.Family] || mutation.ID != expectedMutation.ID || mutation.Family != expectedMutation.Family || mutation.Cell != expectedMutation.Cell || mutation.ChangedSemanticEdge != expectedMutation.ChangedSemanticEdge || mutation.Operator != expectedMutation.Operator || mutation.ExpectedDetector != expectedMutation.ExpectedDetector || !seenCells[mutation.Cell] {
			return fmt.Errorf("mutation %d does not match fixed contract", i+1)
		}
		seenFamilies[mutation.Family] = true
	}
	if source.Authority.RepositoryWrites != 0 || source.Authority.LocalTestExecutions != 0 || source.Authority.CrossProjectRequiredGates != 0 || source.Authority.ProductGenerationAuthorized {
		return fmt.Errorf("authority declaration must be zero and non-escalating")
	}
	return nil
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
