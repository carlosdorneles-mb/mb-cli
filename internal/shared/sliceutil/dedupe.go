package sliceutil

// DedupeStringsPreserveOrder returns a copy of items with duplicate strings removed,
// keeping the first occurrence order.
func DedupeStringsPreserveOrder(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, x := range items {
		if _, ok := seen[x]; ok {
			continue
		}
		seen[x] = struct{}{}
		out = append(out, x)
	}
	return out
}
