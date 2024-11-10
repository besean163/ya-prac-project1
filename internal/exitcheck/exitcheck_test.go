package exitcheck

import (
	"testing"
	"ya-prac-project1/internal/logger"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	logger.Set()
	analysistest.Run(t, analysistest.TestData(), Analyzer, "./...")
}
