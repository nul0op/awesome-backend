package indexer

import (
	"awesome-portal/backend/model"
	"awesome-portal/backend/util"
	"fmt"
	"net/http"
	"strings"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

func Index(link string, depth int, parentCount int) {
	logPrefix := strings.Repeat("  ", depth) + fmt.Sprintf("[%d/%d]", depth, parentCount)

	util.Log.Debugf("%s indexing link: %s", logPrefix, link)
	util.Log.Debugf("%s Getting project metadata", logPrefix)

	metadata, rc := getProjectMetaData(link)

	// FIXME: find a way to save get usefull metadata from website outside GitHub ?
	// FIXME: go get the HEAD of the website and the META tags ?";
	// FIXME: why should i go look at the constructor ?
	// FIXME: get head and use last-modified: Thu, 08 May 2025 08:35:35 GMT (curl --head https://nodejs.org/api/fs.html)
	if rc == model.RC_LINK_HAS_NO_INDEX_PAGE {
		util.Log.Warnf("%s RC_LINK_HAS_NO_INDEX_PAGE: indexing partial data", logPrefix)
		metadata = model.NewAwesomeLink()
		metadata.Description = "Unknown, outside GitHub !"
		metadata.Subscribers = 0
		metadata.Watchers = 0
		metadata.CloneUrl = ""
		metadata.ReadmeUrl = ""
		metadata.OriginUrl = link
		util.Log.Warnf("%s SAVING: %+v", logPrefix, metadata)
		//saveLink(metadata)
		return

	} else if rc == model.RC_LINK_IS_NOT_A_PROJECT_ROOT {
		util.Log.Infof("%s AWESOME: found an individual project outside GitHub. saving it to the index", logPrefix)
		metadata = model.NewAwesomeLink()
		metadata.Description = "Unknown, outside GitHub !"
		metadata.Subscribers = 0
		metadata.Watchers = 0
		metadata.CloneUrl = ""
		metadata.ReadmeUrl = ""
		metadata.OriginUrl = link
		util.Log.Infof("%s SAVING: %+v", logPrefix, metadata)
		//saveLink(metadata)
		return

	} else if rc == model.RC_LINK_IS_A_USER_LANDING_PAGE {
		util.Log.Debugf("%s stopping: user landing page on github ... unable to process", logPrefix)
		return

	} else if rc == model.RC_LINK_IS_NOT_ON_GITHUB {
		// FIXME: try to gather some metadata from external api ? or HEAD last-updated, ...
		// also: look if there is some pub inside the page and drop it if this is the case
		util.Log.Debugf("%s stopping: something outside of github ... unable to process", logPrefix)
		return
	}

	metadata.Level = depth

	headers := make(map[string]string)
	headers["Accept"] = "application/vnd.github.v3.raw"

	rc, content := callGithubAPI(metadata.ReadmeUrl, http.MethodGet, headers)

	if rc != 200 {
		util.Log.Debugf("%s url has no index page. saving page metadata and stopping branch (RC_LINK_HAS_NO_INDEX_PAGE)", logPrefix)
		return
	}

	if metadata.OriginUrl != model.AW_ROOT &&
		!isAwesomeIndexPage(content) {

		// we have reached an leaf, go get the project metadata before closing this branch
		util.Log.Infof("%s AWESOME: found project inside GitHub. saving it to the index", logPrefix)
		util.Log.Infof("%s SAVING: %+v ", logPrefix, metadata)
		return
	}

	util.Log.Debugf("%s looks like an Awesome index, walk over it.", logPrefix)

	content = stripHtmlPrefix(content)
	pLinks := parseMarkdown(content, depth)
	depth++
	// util.Log.Debug(metadata)

	count := 0
	for _, v := range pLinks {
		count++
		logProgress(depth, count, fmt.Sprintf("found link: [%s]. indexinx [%s]", v.Name, v.URL))
		Index(v.URL, depth, count)
	}

}

func logProgress(c1 int, c2 int, s string) {
	prefix := strings.Repeat(" ", c1)
	util.Log.Debugf("%s [%d/%d]: ", prefix, c1, c2)
}

func parseMarkdown(md []byte, depth int) map[string]model.PotentialLink {
	logPrefix := strings.Repeat(" ", depth)

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
			// util.Log.Debugf("trail of node is: %s", trail)
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
				util.Log.Warnf("%s possible duplicate link found: %s", logPrefix, pLink.URL)
			} else {
				pLinks[pLink.URL] = pLink
			}

			// util.Log.Debugf("PARENT : %s => %s", pLink.URL, strings.Join(pp, " > "))
			pLink.Parents = strings.Join(headings, " > ")
		}

		return ast.GoToNext
	})

	return pLinks
	// for _, v := range pLinks {
	// 	// fmt.Printf("LINK FOUND: name: [%s] url: [%s] [%s]\n", v.Name, v.URL, v.Parents)

	// 	// we do not index anchor
	// 	if string(v.URL[0]) == "#" {
	// 		continue
	// 	}
	// 	link := model.AwesomeLink{}
	// 	link.Name = v.Name
	// 	link.Description = v.Description
	// 	link.Topics = v.Parents
	// 	link.OriginUrl = v.URL
	// 	link.OriginHash = util.GetSHA1(v.URL)
	// 	link.UpdateTs = time.Now().Unix()
	// 	link.Subscribers = rand.IntN(4096)
	// 	link.Watchers = rand.IntN(4096)

	// 	if len(model.GetLink(link.OriginHash)) > 0 {
	// 		// FIXME implement update
	// 	} else {
	// 		model.SaveLink(link)
	// 	}
	// }
}
