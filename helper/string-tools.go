package helper

func ContainsString(slice []string, search []string) bool {
	found := 0
	for _, a := range search {
		for _, b := range slice{
			if a == b {
				 found += 1
				break
			}
		}
	}
	if found == len(search) {
		return true
	}
	return false
}
