package utils

import (
	"strings"
)

func WrapErrorMessage(err error, width int) string {
	parts := strings.Split(err.Error(), " ")

	// make the expanded parts as long as parts at first, it might need to be bigger
	// but it should never be smaller
	expandedParts := make([]string, len(parts))

	// loop parts and find extra-long phrases (particularly directory paths)
	// then populate an array where it is all flattened out
	for _, p := range parts {
		// consistent file path separator
		p = strings.ReplaceAll(p, "\\", "/")

		isPath := strings.ContainsAny(p, "/")
		notURL := !strings.Contains(p, "http")

		if notURL && isPath {
			subParts := strings.Split(p, "/")

			for spi, sp := range subParts {
				// if I am not the last one then add a slash
				// the last one needs a space like everything else
				if spi+1 != len(subParts) {
					sp += "/"
				} else {
					sp += " "
				}

				expandedParts = append(expandedParts, sp)
			}
		} else {
			// add the space for word separation
			expandedParts = append(expandedParts, p+" ")
		}
	}

	returnString := strings.Builder{}

	rowWidth := 0
	for _, p := range expandedParts {
		if len(p)+rowWidth+1 <= width {
			rowWidth += len(p) + 1
			returnString.WriteString(p)
		} else {
			returnString.WriteString(p + "\n")
			rowWidth = 0
		}
	}

	return returnString.String()
}
