package slug

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	reNonAlphanumeric = regexp.MustCompile(`[^a-z0-9\-]+`)
	reMultiHyphen     = regexp.MustCompile(`-{2,}`)
)

func Generate(s string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	normalised, _, _ := transform.String(t, s)
	lower := strings.ToLower(normalised)
	lower = strings.NewReplacer(" ", "-", "_", "-").Replace(lower)
	clean := reNonAlphanumeric.ReplaceAllString(lower, "")
	clean = reMultiHyphen.ReplaceAllString(clean, "-")
	return strings.Trim(clean, "-")
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}
