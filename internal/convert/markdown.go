package convert

import (
	"bytes"
	"fmt"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

func MarkdownToHTML(source []byte) (string, error) {
	var buf bytes.Buffer

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	if err := md.Convert(source, &buf); err != nil {
		return "", fmt.Errorf("markdown conversion failed: %w", err)
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<style>
  body { font-family: Arial, sans-serif; }
  code { background-color: #f4f4f4; padding: 2px 4px; font-family: monospace; }
  pre { background-color: #f4f4f4; padding: 10px; overflow-x: auto; }
  pre code { padding: 0; background: none; }
  blockquote { border-left: 3px solid #ccc; margin-left: 0; padding-left: 15px; color: #666; }
</style>
</head>
<body>
%s
</body>
</html>`, buf.String())

	return html, nil
}

func FileToHTML(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return MarkdownToHTML(data)
}
