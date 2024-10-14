package template

import (
	"bytes"
	"golang.org/x/net/html"
	"io"
	"strings"
)

// deprecated
func CompileTemplateDynamicField(body string, data map[string]string) string {
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return body
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i, attr := range n.Attr {
				if strings.ToLower(attr.Key) == "data-dynamic-field" {
					if val := data[attr.Val]; val != "" {

						n.FirstChild = nil
						n.LastChild = nil
						// replace node

						n.AppendChild(&html.Node{
							Type: html.TextNode,
							Data: val,
						})
					}
				}
				n.Attr[i] = attr // must have this or node wont change
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, doc)
	return buf.String()
}

func applyStyle(node *html.Node, style map[string]string) {
	var outstyle string
	for k, val := range style {
		outstyle += ";" + k + ":" + val
	}

	found := false
	for _, attr := range node.Attr {
		if strings.ToLower(attr.Key) == "style" {
			found = true
			break
		}
	}

	if !found {
		node.Attr = append(node.Attr, html.Attribute{Key: "style", Val: ""})
	}

	for i, attr := range node.Attr {
		if strings.ToLower(attr.Key) == "style" {
			attr.Val += outstyle
			node.Attr[i] = attr
		}
	}
}

// inline style
func CompileTemplateToEmail(body string, data map[string]string) string {
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return body
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			cls := ""
			for i, attr := range n.Attr {
				if strings.ToLower(attr.Key) == "class" {
					cls = attr.Val
				}
				if strings.ToLower(attr.Key) == "data-dynamic-field" {
					if val := data[attr.Val]; val != "" {

						n.FirstChild = nil
						n.LastChild = nil
						// replace node

						n.AppendChild(&html.Node{
							Type: html.TextNode,
							Data: val,
						})
					}
				}
				n.Attr[i] = attr // must have this or node wont change
			}

			classes := strings.Split(cls, " ")
			style := map[string]string{}
			for _, cls := range classes {
				if cls == "sbz_lexical_text__bold" {
					style["font-weight"] = "bold"
				}
				if cls == "sbz_lexical_text__underline" {
					style["text-decoration"] = "underline"
				}
				if cls == "sbz_lexical_text__italic" {
					style["font-style"] = "italic"
				}
			}
			if len(style) > 0 {
				applyStyle(n, style)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, doc)
	return buf.String()
}

func CompileTemplateToPlainText(body string) string {
	out := ""
	domDocTest := html.NewTokenizer(strings.NewReader(body))
	previousStartTokenTest := domDocTest.Token()
loopDomTest:
	for {
		tt := domDocTest.Next()
		switch {
		case tt == html.ErrorToken:
			break loopDomTest // End of the document,  done
		case tt == html.StartTagToken:
			previousStartTokenTest = domDocTest.Token()
			if previousStartTokenTest.Data == "p" {
				out += "\n"
			}
			if previousStartTokenTest.Data == "br" {
				out += "\n"
			}
		case tt == html.EndTagToken:
			previousStartTokenTest = html.Token{}
		case tt == html.SelfClosingTagToken:
			if token := domDocTest.Token(); token.Data == "br" {
				out += "\n"
			}
		case tt == html.TextToken:
			if previousStartTokenTest.Data == "script" || previousStartTokenTest.Data == "style" {
				continue
			}
			txt := domDocTest.Text()
			out += html.UnescapeString(string(txt))
		}
	}
	return strings.TrimSpace(out)
}
