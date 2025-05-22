package indexer

import (
	"awesome-portal/backend/model"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gomarkdown/markdown/ast"
)

func getProjectMetaData(url string) (metadata model.AwesomeLink, rc int) {

	model.Log.Debug("Getting project metadata for ", url)

	if strings.Index(url, "https://github.com/") != 0 {
		rc = model.RC_LINK_IS_NOT_ON_GITHUB
		return
	}

	url_parts := strings.Split(url, "/")
	if len(url_parts) > 5 {
		rc = model.RC_LINK_IS_NOT_A_PROJECT_ROOT
		return
	}

	username := url_parts[3]
	repo := url_parts[4]
	headers := make(map[string]string)

	to_try := [2]string{"README.md", "readme.md"}
	for _, page := range to_try {
		model.Log.Debug("trying: ", page, " ... ")
		target := fmt.Sprintf("%s/repos/%s/%s/contents/%s", model.GH_API_URL, username, repo, page)
		rc, _ := fetch(target, http.MethodHead, headers)
		model.Log.Debug(rc)
		if rc == 404 {
			continue
		}
		metadata.ReadmeUrl = target
	}

	if metadata.ReadmeUrl == "" {
		rc = model.RC_LINK_HAS_NO_INDEX_PAGE
		return
	}

	target := fmt.Sprintf("%s/repos/%s/%s", model.GH_API_URL, username, repo)
	headers = make(map[string]string)
	rc, payload := fetch(target, http.MethodGet, headers)

	// 	// err := json.NewDecoder(data).Decode(&gh_response)

	kv := payloadToJson(payload)

	metadata.Name = kv["name"].(string)
	metadata.Description = kv["description"].(string)
	metadata.CloneUrl = kv["clone_url"].(string)
	metadata.OriginUrl = url

	metadata.Subscribers = int(kv["subscribers_count"].(float64))
	metadata.Watchers = int(kv["watchers_count"].(float64))

	date := kv["updated_at"].(string)
	loc, _ := time.LoadLocation("Japan/Tokyo")
	t, err := time.ParseInLocation(time.RFC3339, date, loc)
	if err != nil {
		log.Fatal(err)
	}
	metadata.UpdateTs = t.Unix()

	_topics := kv["topics"].([]interface{})
	topics := []string{}
	for _, v := range _topics {
		// fmt.Println(v)
		topics = append(topics, v.(string))
	}
	model.Log.Debug(strings.Join(topics, ":"))

	// metadata = gh_response
	rc = 0
	return
}

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

// see if the yaml content looks like a AWL Index by looking up the "official" badge regex
func isAwesomeIndexPage(content []byte) bool {
	return strings.Contains(string(content), "https://awesome.re/badge/")
}

// remove everything (html...) up to the first yaml heading tag ('^#')
func stripHtmlPrefix(content []byte) (result []byte) {
	pattern := regexp.MustCompile(`(?m)^#`)
	loc := pattern.FindIndex(content)

	result = content[loc[0]:]
	return
}

// ---------------------------- util for REST api calls

func payloadToJson(data []byte) map[string]interface{} {
	kv := make(map[string]interface{})

	err := json.Unmarshal(data, &kv)
	if err != nil {
		fmt.Printf("%w\n", err)
	}

	// for key, value := range kv {
	// 	fmt.Println(key, value)
	// }

	return kv
}

func fetch(url string, method string, headers map[string]string) (rc int, payload []byte) {
	model.Log.Debug("fetching url: ", url)

	// FIXME: look at the header and throttle when the remote warn us instead
	time.Sleep(1 * time.Second)

	gh_client := http.Client{
		Timeout: time.Second * 5,
	}

	req, error := http.NewRequest(method, url, nil)
	if error != nil {
		model.Log.Error(error)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("GH_TOKEN")))
	for name, value := range headers {
		req.Header.Add(name, value)
		// fmt.Println("HEADER: ", i, v)
	}

	res, error := gh_client.Do(req)
	if error != nil {
		model.Log.Error(error)
	}
	rc = res.StatusCode

	payload, error = io.ReadAll(res.Body)

	if error != nil {
		model.Log.Error(error)
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
