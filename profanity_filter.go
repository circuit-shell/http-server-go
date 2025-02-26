package main

import (
	"strings"
)

func censorProfanity(text string) string {
	// Define our list of profane words
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}

	// Split the text into words
	words := strings.Fields(text)

	// Process each word
	for i, word := range words {
		// Convert word to lowercase for comparison
		wordLower := strings.ToLower(word)

		// Check if the word matches any profane word
		for _, profane := range profaneWords {
			if wordLower == profane {
				words[i] = "****"
				break
			}
		}
	}

	// Join the words back together and return
	return strings.Join(words, " ")
}
