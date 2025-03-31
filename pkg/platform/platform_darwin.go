package platform

func collectInfo() (Info, error) {
	info := Info{
		Id:       "darwin",
		Codename: "ventura, sonoma, (implement me!)",
	}
	return info, nil
}
