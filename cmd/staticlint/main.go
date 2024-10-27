// Package main индивидуалный статический анализатор кода
//
// Содержит:
// 1. Все анализаторы стандартного пакета analisys
// 2. Проверки типа SA пакета statictest
// 3. Проверку на использование os.Exit в пакете main и функции main
package main

import (
	"ya-prac-project1/internal/exitcheck"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stdversion"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	checks := getVetCheckAnalizers()
	checks = append(checks, getStaticCheckAnalizers()...)
	multichecker.Main(checks...)
}

func getVetCheckAnalizers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		exitcheck.Analyzer,
		appends.Analyzer,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		defers.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		httpmux.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		inspect.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		slog.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stdversion.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		usesgenerics.Analyzer,
	}
}

func getStaticCheckAnalizers() []*analysis.Analyzer {
	checks := map[string]bool{
		"SA1000": true,
		"SA1001": true,
		"SA1002": true,
		"SA1003": true,
		"SA1004": true,
		"SA1005": true,
		"SA1006": true,
		"SA1007": true,
		"SA1008": true,
		"SA1010": true,
		"SA1011": true,
		"SA1012": true,
		"SA1013": true,
		"SA1014": true,
		"SA1015": true,
		"SA1016": true,
		"SA1017": true,
		"SA1018": true,
		"SA1019": true,
		"SA1020": true,
		"SA1021": true,
		"SA1023": true,
		"SA1024": true,
		"SA1025": true,
		"SA1026": true,
		"SA1027": true,
		"SA1028": true,
		"SA1029": true,
		"SA1030": true,
		"SA1031": true,
		"SA1032": true,
		"SA2000": true,
		"SA2001": true,
		"SA2002": true,
		"SA2003": true,
		"SA3000": true,
		"SA3001": true,
		"SA4000": true,
		"SA4001": true,
		"SA4003": true,
		"SA4004": true,
		"SA4005": true,
		"SA4006": true,
		"SA4008": true,
		"SA4009": true,
		"SA4010": true,
		"SA4011": true,
		"SA4012": true,
		"SA4013": true,
		"SA4014": true,
		"SA4015": true,
		"SA4016": true,
		"SA4017": true,
		"SA4018": true,
		"SA4019": true,
		"SA4020": true,
		"SA4021": true,
		"SA4022": true,
		"SA4023": true,
		"SA4024": true,
		"SA4025": true,
		"SA4026": true,
		"SA4027": true,
		"SA4028": true,
		"SA4029": true,
		"SA4030": true,
		"SA4031": true,
		"SA4032": true,
		"SA5000": true,
		"SA5001": true,
		"SA5002": true,
		"SA5003": true,
		"SA5004": true,
		"SA5005": true,
		"SA5007": true,
		"SA5008": true,
		"SA5009": true,
		"SA5010": true,
		"SA5011": true,
		"SA5012": true,
		"SA6000": true,
		"SA6001": true,
		"SA6002": true,
		"SA6003": true,
		"SA6005": true,
		"SA6006": true,
		"SA9001": true,
		"SA9002": true,
		"SA9003": true,
		"SA9004": true,
		"SA9005": true,
		"SA9006": true,
		"SA9007": true,
		"SA9008": true,
		"SA9009": true,
		"S1000":  true,
		"ST1000": true,
		"QF1001": true,
	}

	var items []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		// добавляем в массив нужные проверки
		if checks[v.Analyzer.Name] {
			items = append(items, v.Analyzer)
		}
	}
	return items
}
