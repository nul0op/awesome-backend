package indexer

import (
	"awesome-portal/backend/model"
	"fmt"
	"log"
	"net/http"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

func Index(repo string, depth int) {
	fmt.Printf("[%d]: scanning git repository: %s\n", depth, repo)
	metadata, rc := getProjectMetaData("https://github.com/sindresorhus/awesome")

	// FIXME: find a way to save get usefull metadata from website outside GitHub ?
	// FIXME: go get the HEAD of the website and the META tags ?";
	// FIXME: why should i go look at the constructor ?
	// FIXME: get head and use last-modified: Thu, 08 May 2025 08:35:35 GMT (curl --head https://nodejs.org/api/fs.html)
	if rc > 0 {
		fmt.Println("  AWESOME: found an individual project outside GitHub. saving it to the index")
		metadata = model.AwesomeLink{}
		metadata.Description = "Unknown, outside GitHub !"
		metadata.Subscribers = 0
		metadata.Watchers = 0
		metadata.CloneUrl = ""
		metadata.ReadmeUrl = ""
		metadata.OriginUrl = repo
		fmt.Println("  SAVING: ", metadata)
		//saveLink(metadata)
		return
	}

	headers := make(map[string]string)
	headers["Accept"] = "application/vnd.github.v3.raw"

	rc, content := fetch(metadata.ReadmeUrl, http.MethodGet, headers)

	if rc != 200 {
		log.Fatal("ERROR: ", model.RC_LINK_HAS_NO_INDEX_PAGE, ": ", rc)
	}

	if metadata.OriginUrl != model.AW_ROOT &&
		!isAwesomeIndexPage(content) {

		// we have reached an leaf, go get the project metadata before closing this branch
		fmt.Println("  AWESOME: found an individual project inside GitHub. saving it to the index")
		fmt.Println("  SAVING: ", metadata)
		// await saveLink(projectMeta);

		// terminating the branch
		return
	}

	fmt.Println("  looks like an Awesome index, walk over it.")

	content = stripHtmlPrefix(content)
	parseMarkdown(content)

	// fmt.Printf("result is %s\n", result[0:31])
	fmt.Printf("rc is     %d\n", rc)
	fmt.Print(metadata)

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

	ast.WalkFunc(parsed_document, func(node ast.Node, entering bool) ast.WalkStatus {

		// textual content like heading name
		// fmt.Println("CONTENT: ", getContent(node))

		// then type-specific values
		if heading, ok := node.(*ast.Heading); ok && entering {

			prev_level = cur_level
			cur_level = heading.Level
			fmt.Println("debug before: prev_level: ", prev_level, "current: ", cur_level, "heading: ", heading.Level, "content: ", len(headings), ": ", headings)

			// fmt.Print("HEADING: ", getContent(heading))
			if cur_level > prev_level {
				// if we pass through an heading, we push into the heading stack the entry
				fmt.Println("increasing headings")
				// headings = append(headings, getContent(heading))
				headings = append(headings, heading.HeadingID)

			} else if cur_level < prev_level {
				fmt.Println("decresing headings")
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
			fmt.Println("debug after: prev_level: ", prev_level, "current: ", cur_level, "heading: ", heading.Level, "content: ", len(headings), ": ", headings)
		}
		if paragraph, ok := node.(*ast.Paragraph); ok && entering {
			// this is just a container,
			fmt.Println("PARAGRAPH1: ", getFirstChildLabel(paragraph))
			// the first text sibling OR the first link.text.value be used as a heading
			fmt.Println("PARAGRAPH2 ", string("FIXME: go get the first "))

		}
		if link, ok := node.(*ast.Link); ok && entering {
			// fmt.Println("LINK1: ", getContent(link))
			// we break on this line
			fmt.Println("LINK2: ", string(link.Destination))

		}
		if _, ok := node.(*ast.List); ok && entering {
			fmt.Println("LIST: ", getContent(node))
		}
		if _, ok := node.(*ast.ListItem); ok && entering {
			fmt.Println("ITEM: ", getContent(node))
		}
		return ast.GoToNext
	})
}
