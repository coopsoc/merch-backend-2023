package main

func max(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func filter(items []item, fn func(item) bool) string {
	for _, i := range items {
		if fn(i) {
			return i.NAME
		}
	}

	return ""
}
