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
		} else if strings.ContainsAny(s, "/\\") {
			// break on paths too
			for _, pathPart := range strings.Split(strings.ReplaceAll(s, "\\", "/"), "/") {
				if rowWidth+len(pathPart) <= width {
					returnString.WriteString(pathPart + "/")
				} else if len(pathPart) > width {
					// if it is longer than the rest of the line, then write as much as we can
					// before splitting to a new line
					returnString.WriteString(pathPart[:width-rowWidth])
					writeStringChunks(pathPart[width-rowWidth+1:], &returnString, width)
				} else {
					returnString.WriteString("\n" + pathPart)
				}
			}
		} else if len(s) > width {
			// if it is longer than the rest of the line, then write as much as we can
			// before splitting to a new line
			returnString.WriteString(s[:width-rowWidth])
			writeStringChunks(s[width-rowWidth+1:], &returnString, width)
		} else {
			returnString.WriteString("\n" + s)
			rowWidth = len(s)
		}
	}

	return returnString.String()
}

func writeStringChunks(s string, returnString *strings.Builder, width int) {
	// chunk the string, its way too long
	chunks := []strings.Builder{
		// create the initial 0 position builder
		{},
	}

	chunkPosition := 0
	for _, ss := range s {
		chunks[chunkPosition].WriteString(string(ss))
		if chunks[chunkPosition].Len() >= width {
			chunkPosition++
			chunks = append(chunks, strings.Builder{})
		}
	}

	for _, c := range chunks {
		if c.Len() > 0 {
			returnString.WriteString("\n" + c.String())
		}
	}
}
