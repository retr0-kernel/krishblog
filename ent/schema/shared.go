package schema

import "regexp"

// slugRegexp validates URL-safe slugs shared across Post and Section schemas.
var slugRegexp = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
