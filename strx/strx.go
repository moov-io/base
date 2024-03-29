// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package strx

import (
	"cmp"
	"strconv"
	"strings"
)

// Or returns the first non-empty string
func Or(options ...string) string {
	return cmp.Or(options...)
}

// Yes returns true if the provided case-insensitive string matches 'yes' and is used to parse config values.
func Yes(in string) bool {
	in = strings.TrimSpace(in)
	if strings.EqualFold(in, "yes") {
		return true
	}
	v, _ := strconv.ParseBool(in)
	return v
}
