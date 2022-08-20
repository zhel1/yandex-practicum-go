// Package customanalyzers provides custom code analysis.
package customanalyzers

import (
	"golang.org/x/tools/go/analysis/analysistest"
	"testing"
)

func TestMyAnalyzer(t *testing.T) {
	// function analysistest.Run applies OsExitInMainAnalyzer to packages from testdata and checks expected result
	analysistest.Run(t, analysistest.TestData(), OsExitInMainAnalyzer, "./...")
}
