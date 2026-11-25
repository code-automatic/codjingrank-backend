package forum

import (
	"errors"
	"html"
	"net/url"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func safeMarkdown(value string) (string, error) {
	source := []byte(value)
	doc := goldmark.DefaultParser().Parse(text.NewReader(source))
	var raw []text.Segment
	err := ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		var destination []byte
		switch n := node.(type) {
		case *ast.RawHTML:
			for i := 0; i < n.Segments.Len(); i++ {
				raw = append(raw, n.Segments.At(i))
			}
			return ast.WalkContinue, nil
		case *ast.HTMLBlock:
			for i := 0; i < n.Lines().Len(); i++ {
				raw = append(raw, n.Lines().At(i))
			}
			if n.HasClosure() {
				raw = append(raw, n.ClosureLine)
			}
			return ast.WalkContinue, nil
		case *ast.Link:
			destination = n.Destination
		case *ast.Image:
			destination = n.Destination
		default:
			return ast.WalkContinue, nil
		}
		u, err := url.Parse(strings.TrimSpace(html.UnescapeString(string(destination))))
		if err != nil || (u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "mailto") {
			return ast.WalkStop, errors.New("unsafe Markdown destination")
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return "", err
	}
	sort.Slice(raw, func(i, j int) bool { return raw[i].Start > raw[j].Start })
	for _, seg := range raw {
		value = value[:seg.Start] + html.EscapeString(value[seg.Start:seg.Stop]) + value[seg.Stop:]
	}
	return value, nil
}
