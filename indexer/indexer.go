package indexer

import (
	"awesome-portal/backend/model"
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

	pLink := model.PotentialLink{}
	previousName := ""

	pLinks := make(map[string]model.PotentialLink)

	previousListItemParentCount := 0
	listItemParentCount := 0

	trail := [5]string{}
	breaks := "heading-link-paragraph-list-listitem"

	headings := []string{}
	lastSeenHeading := ""

	ast.WalkFunc(parsed_document, func(node ast.Node, entering bool) ast.WalkStatus {

		// we save the last 5 kind of node we see,
		// so last[Ø] is the current, last[1] is the one before, ...
		nt := getNodeType(node)

		if heading, ok := node.(*ast.Heading); ok && entering {
			lastSeenHeading = string(heading.HeadingID)
		}

		if strings.Contains(breaks, nt) {
			trail[4] = trail[3]
			trail[3] = trail[2]
			trail[2] = trail[1]
			trail[1] = trail[0]
			trail[0] = nt
			// model.Log.Debugf("trail of node is: %s", trail)
		}

		// case 1: entering an heading paragraph
		// case 1: IN
		if trail[0] == "list" && trail[1] == "heading" {
			headings = append(headings, lastSeenHeading)
		}

		// case 2: entering a sublist (no heading involved, only spacing)
		// case 2: IN
		if trail[0] == "list" && trail[1] == "paragraph" {
			headings = append(headings, "SUBSTRING: "+previousName)
		}

		// case 2 OUT
		if trail[0] == "listitem" {
			previousListItemParentCount = listItemParentCount
			listItemParentCount = parentCount(node)
			if listItemParentCount < previousListItemParentCount {
				headings = headings[:len(headings)-1]
			}
		}

		// only problem here is the nested list thing
		texts := map[string]string{}
		if link, ok := node.(*ast.Link); ok && entering {
			previousName = pLink.Name
			getParagraphContent(link.GetParent(), &pLink, texts)

			if _, seen := pLinks[pLink.URL]; seen {
				model.Log.Warnf("possible duplicate link found: %s", pLink.URL)
			} else {
				pLinks[pLink.URL] = pLink
			}

			// model.Log.Debugf("PARENT : %s => %s", pLink.URL, strings.Join(pp, " > "))
			pLink.Parents = strings.Join(headings, " > ")
		}

		return ast.GoToNext
	})

	for _, v := range pLinks {
		// fmt.Printf("LINK FOUND: name: [%s] url: [%s] [%s]\n", v.Name, v.URL, v.Parents)

		// we do not index anchor
		if string(v.URL[0]) == "#" {
			continue
		}
		link := model.AwesomeLink{}
		link.Name = v.Name
		link.Description = v.Description
		link.Topics = v.Parents
		link.OriginUrl = v.URL
		model.SaveLinks(link)
	}
}
