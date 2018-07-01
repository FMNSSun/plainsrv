package format

import "io"
import "bufio"
import "strings"
import "html"
import "fmt"

const chunkSize = 8192

const stateNil = 0
const stateInP = 1
const stateInPre = 2
const stateInList = 3
const stateInCode = 4

func writeHtml(w io.Writer, s string) {
	io.WriteString(w, html.EscapeString(s))
}

func writeInP(w io.Writer, line string) {
	forcedBreak := false

	if strings.HasSuffix(line, " ") {
		forcedBreak = true
		line = line[0 : len(line)-1]
	}

	writeHtml(w, line)

	if forcedBreak {
		io.WriteString(w, "<br>")
	}
}

func writeHeading(w io.Writer, line string) {
	var r rune
	lvl := 0
	for _, r = range line {
		if r == '#' {
			lvl++
		}
	}

	if lvl > 5 || lvl < 1 {
		lvl = 5
	}

	if len(line) < (lvl + 1) {
		return
	}

	line = line[lvl+1:]

	io.WriteString(w, fmt.Sprintf("<h%d>", lvl))
	writeHtml(w, line)
	io.WriteString(w, fmt.Sprintf("</h%d>\n", lvl))
}

func Format(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, chunkSize)
	_ = buf
	state := stateNil

	for scanner.Scan() {
		raw := scanner.Bytes()
		line := string(raw)

		if line == "" {
			switch state {
			case stateInP:
				state = stateNil
				io.WriteString(w, "</p>\n")
			case stateInPre:
				state = stateNil
				io.WriteString(w, "</pre>\n")
			case stateInCode:
				state = stateNil
				io.WriteString(w, "</code></pre>\n")
			case stateInList:
				state = stateNil
				io.WriteString(w, "</ul>\n")
			}

			continue
		}

		switch state {
		case stateInList:
			if strings.HasPrefix(line, " * ") {
				line = line[3:]
			} else {
			}

			io.WriteString(w, "  <li>")
			writeHtml(w, line)
			io.WriteString(w, "</li>\n")
		case stateInP:
			w.Write([]byte{32})
			writeInP(w, line)
		case stateInCode:
			if strings.HasPrefix(line, " ) ") {
				line = line[3:]
			} else {
			}

			w.Write([]byte{10})
			writeHtml(w, line)
		case stateInPre:
			if strings.HasPrefix(line, "   ") {
				line = line[3:]
			} else {
			}

			w.Write([]byte{10})
			writeHtml(w, line)
		case stateNil:
			if strings.HasPrefix(line, "   ") {
				state = stateInPre
				w.Write([]byte("<pre>"))
				writeHtml(w, line[3:])
			} else if strings.HasPrefix(line, " ) ") {
				state = stateInCode
				w.Write([]byte("<pre><code>"))
				writeHtml(w, line[3:])
			} else if strings.HasPrefix(line, " * ") {
				state = stateInList
				io.WriteString(w, "<ul>\n")
				io.WriteString(w, "  <li>")
				writeHtml(w, line[3:])
				io.WriteString(w, "</li>\n")
			} else if strings.HasPrefix(line, "#") {
				writeHeading(w, line)
			} else {
				state = stateInP
				w.Write([]byte("<p>"))
				writeInP(w, line)
			}
		}
	}

	return nil
}
