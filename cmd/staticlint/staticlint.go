package main

import (
	"staticlint/noosexit"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/uudashr/iface/unused"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {

	myChecks := []*analysis.Analyzer{
		// Standard analyzer
		atomic.Analyzer,
		loopclosure.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	}

	for _, a := range staticcheck.Analyzers {
		// Staticcheck SA analyzers (all)
		if a.Analyzer.Name[0] == 'S' && a.Analyzer.Name[1] == 'A' {
			myChecks = append(myChecks, a.Analyzer)
		}

		// Additional Staticcheck analyzers: St and S1
		if a.Analyzer.Name == "ST1000" || a.Analyzer.Name == "S1001" {
			myChecks = append(myChecks, a.Analyzer)
		}
	}

	// Stylecheck analyzers
	for _, a := range stylecheck.Analyzers {
		myChecks = append(myChecks, a.Analyzer)
	}

	// Simple analyzers
	for _, a := range simple.Analyzers {
		myChecks = append(myChecks, a.Analyzer)
	}

	// Public analyzers
	myChecks = append(myChecks, errcheck.Analyzer)
	myChecks = append(myChecks, unused.Analyzer)

	// Custom analyzer
	myChecks = append(myChecks, noosexit.Analyzer)

	multichecker.Main(myChecks...)

}
