package hashutil

import (
	"fmt"
	"strings"
)

type CompareReport struct {
	MatchCount            int      `json:"match_count"`
	MismatchCount         int      `json:"mismatch_count"`
	MissingReferenceCount int      `json:"missing_reference_count"`
	MissingActualCount    int      `json:"missing_actual_count"`
	MismatchFiles         []string `json:"mismatch_files,omitempty"`
	MissingReferenceFiles []string `json:"missing_reference_files,omitempty"`
	MissingActualFiles    []string `json:"missing_actual_files,omitempty"`
}

func (r CompareReport) HasFailures() bool {
	return r.MismatchCount > 0 || r.MissingReferenceCount > 0 || r.MissingActualCount > 0
}

// CompareResults compares current hash results against a reference.
func CompareResults(actual, reference []HashResult) CompareReport {
	referenceMap := make(map[string]HashResult)
	for _, ref := range reference {
		referenceMap[comparePathKey(ref.FilePath)] = ref
	}

	actualMap := make(map[string]HashResult)
	report := CompareReport{}

	for _, a := range actual {
		key := comparePathKey(a.FilePath)
		actualMap[key] = a

		ref, found := referenceMap[key]
		if !found {
			report.MissingReferenceFiles = append(report.MissingReferenceFiles, a.FilePath)
			report.MissingReferenceCount++
			continue
		}

		if strings.EqualFold(a.Hash, ref.Hash) {
			report.MatchCount++
		} else {
			report.MismatchFiles = append(report.MismatchFiles, a.FilePath)
			report.MismatchCount++
		}
	}

	for _, ref := range reference {
		key := comparePathKey(ref.FilePath)
		if _, found := actualMap[key]; !found {
			report.MissingActualFiles = append(report.MissingActualFiles, ref.FilePath)
			report.MissingActualCount++
		}
	}

	return report
}

func PrintCompareReport(report CompareReport) {
	fmt.Printf("Summary: %d match, %d mismatch, %d not found in reference, %d missing from actual\n",
		report.MatchCount,
		report.MismatchCount,
		report.MissingReferenceCount,
		report.MissingActualCount,
	)

	if report.MismatchCount > 0 {
		fmt.Println("\nMismatch:")
		for _, path := range report.MismatchFiles {
			fmt.Printf("- %s\n", path)
		}
	}

	if report.MissingReferenceCount > 0 {
		fmt.Println("\nNot found in reference:")
		for _, path := range report.MissingReferenceFiles {
			fmt.Printf("- %s\n", path)
		}
	}

	if report.MissingActualCount > 0 {
		fmt.Println("\nMissing from actual:")
		for _, path := range report.MissingActualFiles {
			fmt.Printf("- %s\n", path)
		}
	}
}
