package asn

import (
	"testing"
)

func TestNormalizeASNName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Test exact keyword matches
		{
			name:     "exact aws match",
			input:    "aws",
			expected: "aws",
		},
		{
			name:     "exact azure match",
			input:    "azure",
			expected: "azure",
		},
		{
			name:     "exact gcp match",
			input:    "gcp",
			expected: "gcp",
		},
		{
			name:     "exact google match",
			input:    "google",
			expected: "gcp",
		},
		{
			name:     "exact yotta match",
			input:    "yotta",
			expected: "yotta",
		},

		// Test case insensitive matching
		{
			name:     "uppercase AWS",
			input:    "AWS",
			expected: "aws",
		},
		{
			name:     "mixed case Azure",
			input:    "Azure",
			expected: "azure",
		},
		{
			name:     "uppercase GOOGLE",
			input:    "GOOGLE",
			expected: "gcp",
		},

		// Test keywords contained in larger strings
		{
			name:     "aws in company name",
			input:    "amazon aws services",
			expected: "aws",
		},
		{
			name:     "azure in company name",
			input:    "microsoft azure cloud",
			expected: "azure",
		},
		{
			name:     "google in company name",
			input:    "google cloud platform",
			expected: "gcp",
		},
		{
			name:     "gcp in company name",
			input:    "gcp infrastructure",
			expected: "gcp",
		},
		{
			name:     "yotta in company name",
			input:    "yotta infrastructure",
			expected: "yotta",
		},

		// Test whitespace handling
		{
			name:     "leading whitespace",
			input:    "  aws",
			expected: "aws",
		},
		{
			name:     "trailing whitespace",
			input:    "azure  ",
			expected: "azure",
		},
		{
			name:     "both leading and trailing whitespace",
			input:    "  google  ",
			expected: "gcp",
		},
		{
			name:     "whitespace with company name",
			input:    "  amazon aws services  ",
			expected: "aws",
		},

		// Test non-matching cases
		{
			name:     "unknown provider",
			input:    "digitalocean",
			expected: "digitalocean",
		},
		{
			name:     "random string",
			input:    "some random provider",
			expected: "some random provider",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
		},

		// Test substring matches (function uses strings.Contains, so these will match)
		{
			name:     "aws as substring",
			input:    "awsome provider",
			expected: "aws", // This will match because "awsome" contains "aws"
		},
		{
			name:     "azure as substring",
			input:    "lazure provider",
			expected: "azure", // This will match because "lazure" contains "azure"
		},
		{
			name:     "gcp as substring",
			input:    "pgcp provider",
			expected: "gcp", // This will match because "pgcp" contains "gcp"
		},

		// Test mixed case with complex strings
		{
			name:     "mixed case complex string",
			input:    "AMAZON Web Services AWS",
			expected: "aws",
		},
		{
			name:     "mixed case with google",
			input:    "Google Cloud Platform GCP",
			expected: "gcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeASNName(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeASNName(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNormalizeASNNameDeterministic tests that the function is deterministic
// for inputs that don't contain multiple keywords
func TestNormalizeASNNameDeterministic(t *testing.T) {
	testCases := []string{
		"aws provider",
		"azure services",
		"google cloud",
		"yotta infrastructure",
		"unknown provider",
	}

	for _, input := range testCases {
		// Run the same input multiple times to ensure deterministic behavior
		results := make(map[string]int)
		for i := 0; i < 10; i++ {
			result := NormalizeASNName(input)
			results[result]++
		}

		if len(results) != 1 {
			t.Errorf("NormalizeASNName(%q) produced non-deterministic results: %v", input, results)
		}
	}
}

// TestNormalizeASNNameMultipleKeywords tests behavior when multiple keywords are present
// Note: Due to Go's map iteration order being non-deterministic, this test verifies
// that the result is one of the valid expected values
func TestNormalizeASNNameMultipleKeywords(t *testing.T) {
	testInput := "aws google azure yotta"

	// The result should be one of the expected normalized names
	validResults := map[string]bool{
		"aws":   true,
		"gcp":   true,
		"azure": true,
		"yotta": true,
	}

	result := NormalizeASNName(testInput)
	if !validResults[result] {
		t.Errorf("NormalizeASNName(%q) = %q, expected one of: aws, gcp, azure, yotta", testInput, result)
	}
}

// TestNormalizeASNNameSpecialCharacters tests the function with special characters
func TestNormalizeASNNameSpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "special characters with aws",
			input:    "aws-services-123",
			expected: "aws",
		},
		{
			name:     "special characters with google",
			input:    "google.cloud",
			expected: "gcp",
		},
		{
			name:     "unicode characters with azure",
			input:    "azure™ services",
			expected: "azure",
		},
		{
			name:     "punctuation only",
			input:    "!@#$%^&*()",
			expected: "!@#$%^&*()",
		},
		{
			name:     "numbers only",
			input:    "12345",
			expected: "12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeASNName(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeASNName(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
