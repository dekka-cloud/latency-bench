package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestNamesStayInsideTheirCells(t *testing.T) {
	var md, csv bytes.Buffer
	rows := []row{{vantage: "Accra | MTN", target: "=HYPERLINK(\"x\")", location: "Lagos\n| injected |", countries: "+GH"}}
	writeMarkdown(&md, rows, 1)
	if !strings.Contains(md.String(), `### From Accra \| MTN`) || strings.Contains(md.String(), "\n| injected") {
		t.Fatalf("markdown: %s", md.String())
	}
	writeCSV(&csv, rows)
	for _, want := range []string{`"'=HYPERLINK(""x"")"`, "'+GH"} {
		if !strings.Contains(csv.String(), want) {
			t.Fatalf("csv lacks %s: %s", want, csv.String())
		}
	}
}
