package utils

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unicode"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateUsername creates a unique username based on the user's name
func GenerateUsername(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Remove non-alphanumeric characters except spaces
	reg := regexp.MustCompile("[^a-z0-9 ]+")
	name = reg.ReplaceAllString(name, "")

	// Replace spaces with empty string to create one continuous string
	name = strings.ReplaceAll(name, " ", "")

	// If name is empty after cleaning, return a generic username
	if len(name) == 0 {
		return generateRandomUsername()
	}

	// Limit length to reasonable size (e.g., 20 characters)
	if len(name) > 20 {
		name = name[:20]
	}

	// If the resulting name is still empty, generate a random one
	if len(name) == 0 {
		return generateRandomUsername()
	}

	return name
}

// generateRandomUsername generates a random username when we can't use the name
func generateRandomUsername() string {
	adjectives := []string{"cool", "super", "awesome", "happy", "smart", "quick", "bright", "bold", "clever", "swift"}
	nouns := []string{"user", "gamer", "player", "explorer", "traveler", "thinker", "creator", "builder", "dreamer", "singer"}

	adj := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]
	number := rand.Intn(10000)

	return fmt.Sprintf("%s%s%d", adj, noun, number)
}

// GenerateUniqueUsername creates a unique username ensuring it doesn't already exist in the database
// This function should be called with access to the user repository to check for existing usernames
func GenerateUniqueUsername(name string, checkAvailable func(username string) bool) string {
	if name == "" {
		return generateRandomUsername()
	}

	baseUsername := GenerateUsername(name)
	username := baseUsername

	// First try the base username
	if checkAvailable(username) {
		return username
	}

	// If not available, try with numbers appended
	attempts := 1
	for attempts <= 100 { // Prevent infinite loop
		username = fmt.Sprintf("%s%d", baseUsername, attempts)
		if checkAvailable(username) {
			return username
		}
		attempts++
	}

	// If all numbered variations fail, return a random username
	return generateRandomUsername()
}

// CleanDisplayName removes non-printable characters and limits length
func CleanDisplayName(displayName string) string {
	var cleaned strings.Builder
	for _, r := range displayName {
		if unicode.IsPrint(r) && cleaned.Len() < 50 { // Limit to 50 chars
			cleaned.WriteRune(r)
		}
	}
	return cleaned.String()
}