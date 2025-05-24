package indexer

import (
	"awesome-portal/backend/model"
	"fmt"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown/ast"
)

func getContent(node ast.Node) string {
	if c := node.AsContainer(); c != nil {
		return contentToString(c.Literal, c.Content)
	}
	leaf := node.AsLeaf()
	return contentToString(leaf.Literal, leaf.Content)
}

func contentToString(d1 []byte, d2 []byte) string {
	if d1 != nil {
		return string(d1)
	}
	if d2 != nil {
		return string(d2)
	}
	return ""
}

// remove everything (html...) up to the first yaml heading tag ('^#')
func stripHtmlPrefix(content []byte) (result []byte) {
	result = content
	pattern := regexp.MustCompile(`(?m)^#`)
	loc := pattern.FindIndex(content)

	if len(loc) > 0 {
		result = content[loc[0]:]
	}
	return
}

// return the number of parent a given node has.
func parentCount(node ast.Node) (count int) {
	if node == nil {
		return 1
	}
	return parentCount(node.GetParent()) + 1
}

// walk up the tree and get all headings and paragraph labels in an ordered list
func getParents(node ast.Node, parents map[string]string) {
	if node == nil {
		return
	}

	getParents(node.GetParent(), parents)

	children := node.GetChildren()
	if len(children) > 0 {
		txt := strings.ToLower(getContent(node.GetChildren()[0]))
		if len(txt) > 0 {
			if _, exists := parents[txt]; exists {
				parents[txt] = txt
			}
		}
	}
}

func getParagraphContent(doc ast.Node, pLink *model.PotentialLink, acc map[string]string) {
	if _, ok := doc.(*ast.Document); ok {
		for _, c := range doc.GetChildren() {
			accumulateChildrenText(c, pLink, acc)
		}
	} else {
		accumulateChildrenText(doc, pLink, acc)
	}

	shortest, longest := "", ""

	for _, v := range acc {
		if len(v) > len(longest) {
			longest = v
			if shortest == "" {
				shortest = longest
			}
		}
		if len(v) < len(shortest) {
			shortest = v
		}
	}
	pLink.Name = normalize(shortest)
	pLink.Description = normalize(longest)

	// if name is missing, tage the last path component from the url (removing any anchor)
	if len(pLink.Name) == 0 {
		url := pLink.URL
		comps := strings.Split(url, "/")
		last := normalize(comps[len(comps)-1])
		regex := regexp.MustCompile(`#.*$`)
		result := regex.ReplaceAllString(last, "")
		pLink.Name = result
	}
}

// accumulate paragraph textual content (top->down)
// lowercase it, removing duplicates along the way
// populate the potential awesomeLink if an url is found somewhere
// depth = 1 means only compute the current node. =x means descend as many as x level down the tree
func accumulateChildrenText(node ast.Node, pLink *model.PotentialLink, acc map[string]string) {
	if node == nil {
		return
	}

	switch v := node.(type) {
	case *ast.Link:
		if pLink != nil {
			pLink.URL = strings.TrimSpace(string(v.Destination))
		}

	case *ast.Text:
		content := strings.ToLower(strings.TrimSpace(getContent(node)))
		if _, seen := acc[content]; !seen {
			acc[content] = content
		}
	}

	for _, child := range node.GetChildren() {
		getParagraphContent(child, pLink, acc)
	}
}

// remove dot/space/tab/hyphen/... at the begining and end
func normalize(s string) (result string) {
	result = ""

	// start of sentence
	regex := regexp.MustCompile(`^[\. \t-]*`)
	result = regex.ReplaceAllString(s, "")

	// end of sentence
	regex = regexp.MustCompile(`[\. \t-]*$`)
	result = regex.ReplaceAllString(result, "")

	return
}

// get a short name of the type of v which excludes package name
// and strips "()" from the end
// in part borrowed from gomarkdown lib since it's not exported by them
func getNodeType(node ast.Node) string {
	s := fmt.Sprintf("%T", node)
	s = strings.ToLower(strings.TrimSuffix(s, "()"))
	if idx := strings.Index(s, "."); idx != -1 {
		return s[idx+1:]
	}
	return s
}

// see if the yaml content looks like a AWL Index by looking up the "official" badge regex
func isAwesomeIndexPage(content []byte) bool {
	return strings.Contains(string(content), "https://awesome.re/badge/") || strings.Contains(string(content), "https://cdn.rawgit.com/sindresorhus/awesome/d7305f38d29fed78fa85652e3a63e154dd8e8829/media/badge.svg")
}
