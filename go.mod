module github.com/marco-m/florist

go 1.26

require (
	github.com/cakturk/go-netstat v0.0.0-20200220111822-e5b49efee7a5
	github.com/creasty/defaults v1.8.0
	github.com/google/go-cmp v0.7.0
	github.com/marco-m/clim v0.1.4
	github.com/marco-m/rosina v0.3.0
)

require github.com/alecthomas/repr v0.5.4 // indirect

retract (
	v0.4.3
	v0.3.1
	v0.3.0
)
