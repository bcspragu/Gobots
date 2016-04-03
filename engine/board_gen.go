package engine

func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
