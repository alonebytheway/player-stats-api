package main

import (
	"testing"
)

func TestCalculatedKD(t *testing.T) {
	tests := []struct {
		name     string
		kills    int
		deaths   int
		expected float64
	}{
		{name: "alonebtw", kills: 10, deaths: 2, expected: 5.0},
		{name: "DenisiniPenisini", kills: 2, deaths: 10, expected: 0.2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateKD(tt.kills, tt.deaths)
			if result != tt.expected {
				t.Errorf("Для %s ожидалось %v, а получили %v", tt.name, tt.expected, result)
			}
		})
	}

}
