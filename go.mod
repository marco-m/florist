module github.com/marco-m/florist

go 1.19

require (
	github.com/alexflint/go-arg v1.4.3
	github.com/cakturk/go-netstat v0.0.0-20200220111822-e5b49efee7a5
	github.com/gertd/wild v0.0.1
	github.com/go-quicktest/qt v0.1.1-0.20221116170248-0c3ea11f9cb5
	github.com/google/go-cmp v0.5.9
	github.com/hashicorp/go-hclog v1.3.1
	github.com/marco-m/xprog v0.3.0
	github.com/scylladb/go-set v1.0.2
)

require (
	github.com/alexflint/go-scalar v1.2.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/rogpeppe/go-internal v1.6.1 // indirect
	golang.org/x/sys v0.1.0 // indirect
	gotest.tools/v3 v3.4.0 // indirect
)

// replace github.com/marco-m/xprog v0.3.0 => ../../xprog

// Update the replacement:
// go mod edit -replace github.com/go-quicktest/qt=github.com/marco-m/qt@main && go mod tidy
replace github.com/go-quicktest/qt => github.com/marco-m/qt v0.0.3-0.20221118180752-89acc9d46294
