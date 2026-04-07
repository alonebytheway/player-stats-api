package main

func CalculateKD(kills int, deaths int) float64 {
	if deaths == 0 {
		return float64(kills)
	} else {
		return float64(kills) / float64(deaths)
	}
}
