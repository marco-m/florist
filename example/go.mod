module example

go 1.27.1

require (
	github.com/marco-m/florist v0.0.0 // Version not used, see "replace" below.
	github.com/marco-m/rosina v0.3.1
)

require (
	github.com/alecthomas/repr v0.5.4 // indirect
	github.com/cakturk/go-netstat v0.0.0-20200220111822-e5b49efee7a5 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/marco-m/clim v0.2.0 // indirect
)

// A replacement without a version on the left side applies to all versions
// of the old module path.
replace github.com/marco-m/florist => ..
