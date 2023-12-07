package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/cockroachdb/cockroachdb-parser/pkg/sql/parser"
	"github.com/cockroachdb/cockroachdb-parser/pkg/sql/sem/tree"
)

var ignoreComments = regexp.MustCompile(`^--.*\s*`)

var cfg = tree.PrettyCfg{
	LineWidth:                60,
	TabWidth:                 4,
	DoNotNewLineAfterColName: false,
	Align:                    tree.PrettyAlignAndDeindent,
	UseTabs:                  true,
	Simplify:                 true,
	Case:                     strings.ToUpper,
	JSONFmt:                  true,
	ValueRedaction:           false,
}

func fmtSQL(stmt string) (string, error) {
	var prettied strings.Builder
	for len(stmt) > 0 {
		stmt = strings.TrimSpace(stmt)
		hasContent := false
		// Trim comments, preserving whitespace after them.
		for {
			found := ignoreComments.FindString(stmt)
			if found == "" {
				break
			}
			// Remove trailing whitespace but keep up to 2 newlines.
			prettied.WriteString(strings.TrimRightFunc(found, unicode.IsSpace))
			newlines := strings.Count(found, "\n")
			if newlines > 2 {
				newlines = 2
			}
			prettied.WriteString(strings.Repeat("\n", newlines))
			stmt = stmt[len(found):]
			hasContent = true
		}
		// Split by semicolons
		next := stmt
		if pos, _ := parser.SplitFirstStatement(stmt); pos > 0 {
			next = stmt[:pos]
			stmt = stmt[pos:]
		} else {
			stmt = ""
		}
		// This should only return 0 or 1 responses.
		allParsed, err := parser.Parse(next)
		if err != nil {
			return "", fmt.Errorf("error parsing %q: %v", next, err)
		}

		for _, parsed := range allParsed {
			prettied.WriteString(cfg.Pretty(parsed.AST))
			prettied.WriteString(";\n")
			hasContent = true
		}
		if hasContent {
			prettied.WriteString("\n")
		}
	}

	return strings.TrimRightFunc(prettied.String(), unicode.IsSpace), nil
}

func main() {
	rawSql, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	out, err := fmtSQL(string(rawSql))
	if err != nil {
		panic(err)
	}

	fmt.Println(out)
}
