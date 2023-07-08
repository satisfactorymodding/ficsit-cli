package utils

import "strings"

func WrapErrorMessage(err error, width int) string {
	parts := strings.Split(err.Error(), " ")

	returnString := strings.Builder{}

	rowWidth := 0
	for _, s := range parts {
		if rowWidth+len(s) <= width {
			returnString.WriteString(s + " ")
			rowWidth += len(s) + 1
		} else {
			returnString.WriteString(s + "\n")
			rowWidth = 0
		}
	}

	return returnString.String()
}
