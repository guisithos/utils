package rand

import (
	"strings"
	"testing"
)

func TestNumber(t *testing.T) {
	const iterations = 1000
	seen := make(map[int64]bool)

	for i := 0; i < iterations; i++ {
		n, err := Number()
		if err != nil {
			t.Errorf("Number() error = %v", err)
			return
		}

		// Track unique numbers
		seen[n] = true
	}

	// With true randomness, we expect a high percentage of unique numbers
	// Given the massive range of int64, getting even 2 duplicates would be extremely unlikely
	uniqueRatio := float64(len(seen)) / float64(iterations)
	if uniqueRatio < 0.99 {
		t.Errorf("Expected mostly unique numbers, but got uniqueness ratio of %v", uniqueRatio)
	}
}

func TestNumberInRange(t *testing.T) {
	tests := []struct {
		name    string
		min     int64
		max     int64
		wantErr bool
	}{
		{
			name:    "success - valid range",
			min:     1,
			max:     100,
			wantErr: false,
		},
		{
			name:    "success - same min and max",
			min:     5,
			max:     5,
			wantErr: false,
		},
		{
			name:    "fail - invalid range",
			min:     100,
			max:     1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NumberInRange(tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("NumberInRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if got < tt.min || got > tt.max {
					t.Errorf("NumberInRange() = %v, want between %v and %v", got, tt.min, tt.max)
				}
			}
		})
	}
}

func TestString(t *testing.T) {
	s1, err := String()
	if err != nil {
		t.Errorf("String() error = %v", err)
		return
	}

	s2, err := String()
	if err != nil {
		t.Errorf("String() error = %v", err)
		return
	}

	if len(s1) != DefaultLength {
		t.Errorf("String() length = %v, want %v", len(s1), DefaultLength)
	}

	if s1 == s2 {
		t.Errorf("Generated strings are equal: %v", s1)
	}
}

func TestStringWithLength(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantLen int
		wantErr bool
	}{
		{
			name:    "success - positive length",
			length:  15,
			wantLen: 15,
			wantErr: false,
		},
		{
			name:    "success - zero length",
			length:  0,
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "fail - negative length",
			length:  -1,
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StringWithLength(tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringWithLength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && len(got) != tt.wantLen {
				t.Errorf("StringWithLength() length = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestPick(t *testing.T) {
	tests := []struct {
		name    string
		slice   []string
		wantErr bool
	}{
		{
			name:    "success - non-empty slice",
			slice:   []string{"a", "b", "c"},
			wantErr: false,
		},
		{
			name:    "fail - empty slice",
			slice:   []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Pick(tt.slice)
			if (err != nil) != tt.wantErr {
				t.Errorf("Pick() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && !contains(tt.slice, got) {
				t.Errorf("Pick() returned value %v not found in slice", got)
			}
		})
	}
}

func TestShuffle(t *testing.T) {
	// Create a larger test slice
	original := make([]int, 100)
	for i := range original {
		original[i] = i
	}

	// Run multiple iterations to ensure consistent behavior
	const iterations = 1000
	positionCounts := make([]map[int]int, len(original))
	for i := range positionCounts {
		positionCounts[i] = make(map[int]int)
	}

	for iter := 0; iter < iterations; iter++ {
		shuffled := make([]int, len(original))
		copy(shuffled, original)

		err := Shuffle(shuffled)
		if err != nil {
			t.Errorf("Shuffle() error = %v", err)
			return
		}

		// Check length hasn't changed
		if len(shuffled) != len(original) {
			t.Errorf("Shuffle() changed slice length")
			return
		}

		// Track positions of each number to verify distribution
		for pos, val := range shuffled {
			positionCounts[val][pos]++
		}

		// Verify elements are still presente
		seen := make(map[int]bool)
		for _, v := range shuffled {
			seen[v] = true
		}
		for _, v := range original {
			if !seen[v] {
				t.Errorf("Shuffle() lost element %v", v)
				return
			}
		}
	}

	// Check distribution of positions
	// Each number should appear in each position roughly iterations/len(original) times
	expectedPerPosition := float64(iterations) / float64(len(original))
	tolerance := expectedPerPosition * 0.5 // Allow 50% deviation

	for num, positions := range positionCounts {
		for pos, count := range positions {
			if float64(count) < expectedPerPosition-tolerance || float64(count) > expectedPerPosition+tolerance {
				t.Errorf("Number %v appeared in position %v %d times, expected roughly %.2f (±%.2f)",
					num, pos, count, expectedPerPosition, tolerance)
			}
		}
	}
}

func TestStringWithCharset(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		charset string
		wantLen int
		wantErr bool
	}{
		{
			name:    "success - numeric only charset",
			length:  10,
			charset: "0123456789",
			wantLen: 10,
			wantErr: false,
		},
		{
			name:    "fail - empty charset",
			length:  10,
			charset: "",
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "fail - negative length",
			length:  -1,
			charset: "abc",
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StringWithCharset(tt.length, tt.charset)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringWithCharset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if len(got) != tt.wantLen {
					t.Errorf("StringWithCharset() length = %v, want %v", len(got), tt.wantLen)
				}

				// Verify all characters are from the charset
				for _, c := range got {
					if !strings.ContainsRune(tt.charset, c) {
						t.Errorf("StringWithCharset() contains character %c not in charset", c)
					}
				}
			}
		})
	}
}

// TestPick helper
func contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
