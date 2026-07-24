package util

import "testing"

func TestCompareSemverPrereleaseIdentifiers(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{"numeric prerelease comparison", "1.0.0-beta.11", "1.0.0-beta.2", 1},
		{"numeric identifiers sort before strings", "1.0.0-beta.1", "1.0.0-beta.alpha", -1},
		{"longer prerelease sorts later", "1.0.0-alpha", "1.0.0-alpha.1", -1},
		{"release sorts after prerelease", "1.0.0-rc.1", "1.0.0", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareSemver(tt.a, tt.b); got != tt.want {
				t.Fatalf("compareSemver(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
