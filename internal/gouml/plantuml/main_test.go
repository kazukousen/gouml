package plantuml_test

func trim(src string) string {
	dst := make([]byte, 0, len(src))
	for _, ch := range src {
		if ch == '\t' {
			continue
		}
		dst = append(dst, byte(ch))
	}
	return string(dst)
}
