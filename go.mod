module github.com/marco-m/florist

go 1.24

require (
	github.com/cakturk/go-netstat v0.0.0-20200220111822-e5b49efee7a5
	github.com/creasty/defaults v1.8.0
	github.com/marco-m/clim v0.1.3-0.20241017082646-199c5f45aa8f
	github.com/marco-m/rosina v0.1.2
	github.com/rogpeppe/go-internal v1.14.1
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/tools v0.32.0 // indirect
)

retract (
	v0.4.3
	v0.3.1
	v0.3.0
)
