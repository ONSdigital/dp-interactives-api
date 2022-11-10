package data

import (
	"regexp"
	"strings"
	"unicode"

	gonanoid "github.com/matoous/go-nanoid/v2"
	uuid "github.com/satori/go.uuid"
)

type Generator func(string) string

var (
	englishArticlesRegEx = regexp.MustCompile("\\b(a|an|and|the)\\b")
	whitespaceRegEx      = regexp.MustCompile("\\s+")
	resourceIdAlphabet   = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" //default includes [-_]
)

// GenerateResourceId should return "short, unique, non-incremental ID" - defined as: [A-Za-z0-9]{8}
func GenerateResourceId() func(string) string {
	return func(string) string {
		return gonanoid.MustGenerate(resourceIdAlphabet, 8)
	}
}

// GenerateHumanReadableSlug will return a slug for given title - defined as:
// 	  "Human readable slug is a short hyphenated slug that aligns with the title of the page, is clear and unambiguous, but is free from articles (a, an, the) and other superfluous words"
// It can be edited manually removing superflous words before published.
func GenerateHumanReadableSlug() func(string) string {
	return func(title string) string {
		var stripped []rune
		for _, c := range title {
			if unicode.IsLetter(c) || unicode.IsDigit(c) {
				stripped = append(stripped, unicode.ToLower(c))
			} else if unicode.IsSpace(c) || c == '-' {
				stripped = append(stripped, ' ')
			}
		}
		withoutArticles := englishArticlesRegEx.ReplaceAllString(string(stripped), "")
		return whitespaceRegEx.ReplaceAllString(strings.TrimSpace(withoutArticles), "-")
	}
}

func GenerateUUID() func(string) string {
	return func(string) string {
		if uid, err := uuid.NewV4(); err == nil {
			return uid.String()
		}
		return ""
	}
}
