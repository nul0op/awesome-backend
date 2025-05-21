package indexer

import (
	"awesome-portal/backend/model"
	"fmt"
	"net/http"

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

	pLink := model.PotentialLink{}
	pLinks := make(map[string]model.PotentialLink)

	ast.WalkFunc(parsed_document, func(node ast.Node, entering bool) ast.WalkStatus {
		if _, ok := node.(*ast.Heading); ok && entering {

			// fmt.Print("HEADING: ", getContent(heading))
			// if cur_level > prev_level {
			// 	// if we pass through an heading, we push into the heading stack the entry
			// 	model.Log.Debug("increasing headings")
			// 	// headings = append(headings, getContent(heading))
			// 	headings = append(headings, heading.HeadingID)

			// } else if cur_level < prev_level {
			// 	model.Log.Debug("decresing headings")
			// 	if len(headings) > 0 {
			// 		headings = headings[:len(headings)-1]
			// 	}

			// } else if cur_level == prev_level {
			// 	if len(headings) > 0 {
			// 		headings = headings[:len(headings)-1]
			// 	}
			// 	headings = append(headings, heading.HeadingID)
			// replace the last
			// headings[heading.Level-1] = getContent(heading)

			// } else {
			// 	log.Panic("Oops.. mismatch in heading", len(headings), " VS ", heading.Level)
		}
		// }

		texts := map[string]string{}
		if link, ok := node.(*ast.Link); ok && entering {
			GetParagraphContent(link.GetParent(), &pLink, texts)

			if _, hasBeenSeen := pLinks[pLink.URL]; hasBeenSeen {
				model.Log.Warnf("possible duplicate link found: %s", pLink.URL)
			} else {
				pLinks[pLink.URL] = pLink
			}

		}

		// if link, ok := node.(*ast.Link); ok && entering {
		// 	pLink.URL = strings.ToLower(string(link.Destination))
		// 	pLink.Name = GetSiblingTextContent(link)
		// 	fmt.Println(ast.ToString(link.GetParent()))
		// 	// parents := []string{}
		// 	// getParents(link, &parents)

		// 	if _, hasBeenSeen := pLinks[pLink.URL]; hasBeenSeen {
		// 		model.Log.Warn("possible duplicate link found !: %s", pLink.URL)
		// 	} else {
		// 		pLinks[pLink.URL] = pLink
		// 	}
		// }

		return ast.GoToNext
	})

	// FIXME: remove anchor, ...
	for _, v := range pLinks {
		fmt.Printf("LINK FOUND: name: [%s] description: [%s] url: [%s]\n", v.Name, v.Description, v.URL)
	}
}

// walk up the tree and get all textual content in an ordered list
// take into accounts link label and headings
func getParents(node ast.Node, parents *[]string) {
	if node == nil {
		return
	}

	// recursively walks up
	fmt.Println(ast.ToString(node))

	// model.Log.Debug("PARENT: walking up: ")
	// if heading, ok := node.(*ast.Heading); ok {
	// 	fmt.Println("PARENT: found heading: ", heading.HeadingID)
	// }
	// if list, ok := node.(*ast.List); ok {
	// 	fmt.Println("LIST: ", list.ID)
	// }
	// if listItem, ok := node.(*ast.ListItem); ok {
	// 	fmt.Println("ITEM: ", listItem.ID)
	// }

	getParents(node.GetParent(), parents)

	content := getContent(node)

	if len(content) > 0 {
		model.Log.Debug("PARENT: ", content)
		//parents = append(parents, content)
	}
}
