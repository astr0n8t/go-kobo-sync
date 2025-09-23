package main

import (
	"strings"
	"testing"
	"unicode"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple filename",
			input:    "Book Title",
			expected: "Book_Title",
		},
		{
			name:     "Filename with problematic characters",
			input:    "Book/Title\\With:Many*Problems?\"<>|",
			expected: "Book_Title_With_Many_Problems",
		},
		{
			name:     "Filename with spaces and dots",
			input:    "  Book Title  .",
			expected: "Book_Title",
		},
		{
			name:     "Empty filename",
			input:    "",
			expected: "unknown",
		},
		{
			name:     "Only spaces and dots",
			input:    "  ...  ",
			expected: "unknown",
		},
		{
			name:     "Non-ASCII characters",
			input:    "Book Título with émojis 😀",
			expected: "Book_Titulo_with_emojis",
		},
		{
			name:     "Very long filename",
			input:    strings.Repeat("Very Long Book Title ", 10),
			expected: "Very_Long_Book_Title_Very_Long_Book_Title_Very_Long_Book_Title_Very_Long_Book_Title_Very_Long_Book_Title_Very_Long_Book_Title_Ve",
		},
		{
			name:     "Exactly 128 characters",
			input:    strings.Repeat("a", 128),
			expected: strings.Repeat("a", 128),
		},
		{
			name:     "More than 128 characters",
			input:    strings.Repeat("a", 150),
			expected: strings.Repeat("a", 128),
		},
		{
			name:     "Mixed ASCII and non-ASCII under 128 chars",
			input:    "Normal Title with Special Chars àáâã",
			expected: "Normal_Title_with_Special_Chars_aaaa",
		},
		{
			name:     "Filename ending with underscores after truncation",
			input:    strings.Repeat("a", 125) + "____",
			expected: strings.Repeat("a", 125),
		},
		{
			name:     "Only problematic characters",
			input:    "/\\:*?\"<>|",
			expected: "unknown",
		},
		{
			name:     "Leading and trailing underscores",
			input:    "_Book Title_",
			expected: "Book_Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			
			// Check length constraint
			if len(result) > 128 {
				t.Errorf("sanitizeFilename() result too long = %d chars, want <= 128", len(result))
			}
			
			// Check ASCII only
			for _, r := range result {
				if r > 127 {
					t.Errorf("sanitizeFilename() contains non-ASCII character: %c", r)
				}
			}
			
			// Check no spaces
			if strings.Contains(result, " ") {
				t.Errorf("sanitizeFilename() contains spaces, should use underscores")
			}
			
			// Check expected result
			if result != tt.expected {
				t.Errorf("sanitizeFilename() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test the function integration in the WebDAV context
func TestWebDAVFilenameGeneration(t *testing.T) {
	testCases := []struct {
		bookTitle    string
		expectedFile string
	}{
		{
			bookTitle:    "The Great Gatsby",
			expectedFile: "The_Great_Gatsby.md",
		},
		{
			bookTitle:    "Война и мир", // War and Peace in Russian
			expectedFile: "unknown.md",
		},
		{
			bookTitle:    "Les Misérables",
			expectedFile: "Les_Miserables.md",
		},
		{
			bookTitle:    strings.Repeat("Very Long Book Title ", 10),
			expectedFile: "Very_Long_Book_Title_Very_Long_Book_Title_Very_Long_Book_Title_Very_Long_Book_Title_Very_Long_Book_Title_Very_Long_Book_Title_Ve.md",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.bookTitle, func(t *testing.T) {
			filename := sanitizeFilename(tc.bookTitle) + ".md"
			if filename != tc.expectedFile {
				t.Errorf("Expected filename %s, got %s", tc.expectedFile, filename)
			}
			
			// Ensure the full filename is still under reasonable length
			if len(filename) > 131 { // 128 + ".md"
				t.Errorf("Full filename too long: %d chars", len(filename))
			}
		})
	}
}

// Helper function to check if a string is ASCII only
func isASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}