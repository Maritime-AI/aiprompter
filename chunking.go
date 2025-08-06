package aiprompter

import "unicode/utf8"

// ChunkTextByMaxBytes splits a string into multiple UTF-8 safe chunks,
// each not exceeding maxBytes in length.
func ChunkTextByMaxBytes(text string, maxBytes int) []string {
	if maxBytes <= 0 || len(text) == 0 {
		return nil
	}

	var chunks []string
	bytes := []byte(text)
	totalLen := len(bytes)
	start := 0

	for start < totalLen {
		end := start + maxBytes
		if end > totalLen {
			end = totalLen
		} else {
			// Backtrack to a valid UTF-8 rune boundary
			for end > start && !utf8.RuneStart(bytes[end]) {
				end--
			}
			// Fallback to force progress if boundary not found
			if end == start {
				end = start + maxBytes
			}
		}

		chunks = append(chunks, string(bytes[start:end]))
		start = end
	}

	return chunks
}
