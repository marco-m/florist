// Some ideas are taken from
// https://github.com/python-distro/distro
// https://github.com/kdeldycke/extra-platforms
package platform

type Info struct {
	// macos
	// ubuntu
	// debian
	// freebsd
	// ...
	Id string

	Codename string
}

// CollectInfo returns information about the current platform.
func CollectInfo() (Info, error) {
	return collectInfo()
}
