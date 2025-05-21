package indexer

import (
	"awesome-portal/backend/model"
	"fmt"
	"net/http"
	"strings"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

func Index(repo string, depth int) {
	model.Log.Debug("[%d]: scanning git repository: %s\n", depth, repo)
	metadata, rc := getProjectMetaData("https://github.com/sindresorhus/awesome")

	// FIXME: find a way to save get usefull metadata from website outside GitHub ?
	// FIXME: go get the HEAD of the website and the META tags ?";
	// FIXME: why should i go look at the constructor ?
	// FIXME: get head and use last-modified: Thu, 08 May 2025 08:35:35 GMT (curl --head https://nodejs.org/api/fs.html)
	if rc > 0 {
		model.Log.Debug("  AWESOME: found an individual project outside GitHub. saving it to the index")
		metadata = model.AwesomeLink{}
		metadata.Description = "Unknown, outside GitHub !"
		metadata.Subscribers = 0
		metadata.Watchers = 0
		metadata.CloneUrl = ""
		metadata.ReadmeUrl = ""
		metadata.OriginUrl = repo
		model.Log.Debug("  SAVING: ", metadata)
		//saveLink(metadata)
		return
	}

	headers := make(map[string]string)
	headers["Accept"] = "application/vnd.github.v3.raw"

	rc, content := fetch(metadata.ReadmeUrl, http.MethodGet, headers)

	if rc != 200 {
		model.Log.Error("ERROR: ", model.RC_LINK_HAS_NO_INDEX_PAGE, ": ", rc)
	}

	if metadata.OriginUrl != model.AW_ROOT &&
		!isAwesomeIndexPage(content) {

		// we have reached an leaf, go get the project metadata before closing this branch
		model.Log.Debug("  AWESOME: found an individual project inside GitHub. saving it to the index")
		model.Log.Debug("  SAVING: ", metadata)
		// await saveLink(projectMeta);

		// terminating the branch
		return
	}

	model.Log.Debug("  looks like an Awesome index, walk over it.")

	content = stripHtmlPrefix(content)
	parseMarkdown(content)

	// fmt.Printf("result is %s\n", result[0:31])
	model.Log.Debug("rc is     %d\n", rc)
	model.Log.Debug(metadata)

	// 	console.debug(`  no index page found (${error.message})`);
	// return;
}

func parseMarkdown(md []byte) {

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	parsed_document := p.Parse(md)

	// OK to dump the tree
	// fmt.Println(ast.ToString(parsed_document))

	// the "algorithm" here is to pass through each node, from top to bottom of the page
	// and saving heading/depth allong the way.
	// each time we see a new link, we create a new record with all found headings and the link
	// each found link at a given level will override the previous in the context
	// also, if a hit a paragrap, thake the first "link"ink has some child in the tree, duplicate and also use it as a heading
	headings := []string{}
	prev_level := -1
	cur_level := -1

	pLink := model.PotentialLink{}
	pLinks := make(map[string]model.PotentialLink)

	ast.WalkFunc(parsed_document, func(node ast.Node, entering bool) ast.WalkStatus {

		// textual content like heading name
		// fmt.Println("CONTENT: ", getContent(node))

		// then type-specific values
		if heading, ok := node.(*ast.Heading); ok && entering {

			prev_level = cur_level
			cur_level = heading.Level
			model.Log.Debugf("before: prev: %d cur: %d heading: %d len: %d: %s", prev_level, cur_level, heading.Level, len(headings), headings)

			// fmt.Print("HEADING: ", getContent(heading))
			if cur_level > prev_level {
				// if we pass through an heading, we push into the heading stack the entry
				model.Log.Debug("increasing headings")
				// headings = append(headings, getContent(heading))
				headings = append(headings, heading.HeadingID)

			} else if cur_level < prev_level {
				model.Log.Debug("decresing headings")
				if len(headings) > 0 {
					headings = headings[:len(headings)-1]
				}

			} else if cur_level == prev_level {
				if len(headings) > 0 {
					headings = headings[:len(headings)-1]
				}
				headings = append(headings, heading.HeadingID)
				// replace the last
				// headings[heading.Level-1] = getContent(heading)

				// } else {
				// 	log.Panic("Oops.. mismatch in heading", len(headings), " VS ", heading.Level)
			}

			// fmt.Println(prev_level, " VS ", cur_level)
			model.Log.Debugf("after: prev: %d cur: %d heading: %d len: %d: %s", prev_level, cur_level, heading.Level, len(headings), headings)
			model.Log.Debug("--")
		}
		// if paragraph, ok := node.(*ast.Paragraph); ok && entering {
		// 	// this is just a container,
		// 	model.Log.Debug("PARAGRAPH1: ", getFirstChildLabel(paragraph))
		// 	// the first text sibling OR the first link.text.value be used as a heading
		// 	model.Log.Debug("PARAGRAPH2 ", string("FIXME: go get the first "))

		// }

		if link, ok := node.(*ast.Link); ok && entering {
			pLink.URL = strings.ToLower(string(link.Destination))
			pLink.Name = GetSiblingTextContent(link)

			if _, hasBeenSeen := pLinks[pLink.URL]; hasBeenSeen {
				model.Log.Warn("possible duplicate link found !: %s", pLink.URL)
			} else {
				pLinks[pLink.URL] = pLink
			}
		}
		// if _, ok := node.(*ast.List); ok && entering {
		// 	model.Log.Debug("LIST: ", getContent(node))
		// }
		// if _, ok := node.(*ast.ListItem); ok && entering {
		// 	model.Log.Debug("ITEM: ", getContent(node))
		// }
		return ast.GoToNext
	})

	// FIXME: remove anchor, ...
	for _, v := range pLinks {
		fmt.Println("LINK FOUND: ", v)
	}
}

// func GetChildrenTextContent(node ast.Node) (result string) {
// 	texts := []string{}

// 	for _, _ = range node.GetChildren() {
// 		text := getContent(node.GetChildren()[0])
// 		if len(text) > 0 {
// 			texts = append(texts, text)
// 		}
// 	}

// 	return strings.Join(texts, ";")
// }

// FIXME: multiple formatting can be set (and thus multiple children), i guess we should go deeper to fetch all the content
// beware of the case were there is link below => we should get out if we find a paragraph.
func GetSiblingTextContent(node ast.Node) string {
	texts := make(map[string]string)

	for _, _ = range node.GetParent().GetChildren() {
		text := getContent(node.GetChildren()[0])
		if len(text) == 0 {
			continue
		}
		if _, hasBeenSeen := texts[text]; !hasBeenSeen {
			texts[text] = text
		}
	}

	results := []string{}
	for _, v := range texts {
		results = append(results, v)
	}

	return strings.Join(results, ";")
}
