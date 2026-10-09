package env

import "testing"

func TestParseSizeReadsDockerStatsUnits(t *testing.T) {
	cases := map[string]float64{
		"1.5GiB": 1.5 * (1 << 30),
		"512MiB": 512 * (1 << 20),
		"900KiB": 900 * (1 << 10),
		"0B":     0,
		"2.5MB":  2.5e6,
	}
	for in, want := range cases {
		got, err := parseSize(in)
		if err != nil || got != want {
			t.Errorf("parseSize(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	if _, err := parseSize("12 parsecs"); err == nil {
		t.Error("an unknown unit must be an error")
	}
}
