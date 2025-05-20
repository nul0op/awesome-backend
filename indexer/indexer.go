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
	"github.com/gomarkdown/markdown/parser"
)

// link and it's metadata found while scanning an Awesome Index page (readme).
// it can become an AwesomeLink at some point
type PotentialLink struct {
	parents []string
	name    string
	url     string
}

// act as a constructor for struct
func NewPotentialLink() PotentialLink {
	self := PotentialLink{}
	self.parents = []string{}
	self.name = ""
	self.url = ""
	return self
}

type AwesomeLink struct {
	// Name()	string
	// SetName( name string)

	name        string
	description string
	level       uint
	update_ts   int64
	origin_url  string // first URL found that defines this link
	index_url   string
	clone_url   string
	origin_hash string //sha256 hash of origin, used to "garante" unicity inside the index
	watchers    float64
	subscribers float64
	topics      []string
}

// act as a constructor for struct
func NewAwesomeLink() AwesomeLink {
	self := AwesomeLink{}
	self.index_url = ""
	return self
}

type GHResponse struct {
	name      string `json:"string"`
	full_name string `json:"string"`
}

func GetProjectMetaData(url string) (metadata AwesomeLink, rc int) {

	fmt.Println("Getting project metadata for ", url)

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
		fmt.Print("trying: ", page, " ... ")
		target := fmt.Sprintf("%s/repos/%s/%s/contents/%s", model.GH_API_URL, username, repo, page)
		rc, _ := throttled_fetch(target, http.MethodHead, headers)
		fmt.Println(rc)
		if rc == 404 {
			continue
		}
		metadata.index_url = target
	}

	if metadata.index_url == "" {
		rc = model.RC_LINK_HAS_NO_INDEX_PAGE
		return
	}

	target := fmt.Sprintf("%s/repos/%s/%s", model.GH_API_URL, username, repo)
	headers = make(map[string]string)
	rc, payload := throttled_fetch(target, http.MethodGet, headers)

	// 	// err := json.NewDecoder(data).Decode(&gh_response)

	kv := payload_to_json(payload)

	metadata.name = kv["name"].(string)
	metadata.description = kv["description"].(string)
	metadata.clone_url = kv["clone_url"].(string)
	metadata.origin_url = url

	metadata.subscribers = kv["subscribers_count"].(float64)
	metadata.watchers = kv["watchers_count"].(float64)

	date := kv["updated_at"].(string)
	loc, _ := time.LoadLocation("Japan/Tokyo")
	t, err := time.ParseInLocation(time.RFC3339, date, loc)
	if err != nil {
		log.Fatal(err)
	}
	metadata.update_ts = t.Unix()

	_topics := kv["topics"].([]interface{})
	topics := []string{}
	for _, v := range _topics {
		// fmt.Println(v)
		topics = append(topics, v.(string))
	}
	fmt.Println(strings.Join(topics, ":"))

	// metadata = gh_response
	rc = 0
	return
}

func throttled_fetch(url string, method string, headers map[string]string) (rc int, payload []byte) {
	log.Println("DEBUG: fetching url: ", url)

	time.Sleep(1 * time.Second)

	gh_client := http.Client{
		Timeout: time.Second * 5,
	}

	req, error := http.NewRequest(method, url, nil)
	if error != nil {
		log.Fatal(error)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("GH_TOKEN")))
	for name, value := range headers {
		req.Header.Add(name, value)
		// fmt.Println("HEADER: ", i, v)
	}

	res, error := gh_client.Do(req)
	if error != nil {
		log.Fatal(error)
	}
	rc = res.StatusCode

	payload, error = io.ReadAll(res.Body)

	if error != nil {
		log.Fatal(error)
	}

	return
}

func payload_to_json(data []byte) map[string]interface{} {
	kv := make(map[string]interface{})

	err := json.Unmarshal(data, &kv)
	if err != nil {
		log.Fatal(err)
	}

	// for key, value := range kv {
	// 	fmt.Println(key, value)
	// }

	return kv
}

func dump_headers(r *http.Response) {
	for name, values := range r.Header {
		for _, value := range values {
			fmt.Println(name, value)
		}
	}
}

func Index(repo string, depth int) {
	fmt.Printf("[%d]: scanning git repository: %s\n", depth, repo)
	metadata, rc := GetProjectMetaData("https://github.com/sindresorhus/awesome")

	// FIXME: find a way to save get usefull metadata from website outside GitHub ?
	// FIXME: go get the HEAD of the website and the META tags ?";
	// FIXME: why should i go look at the constructor ?
	// FIXME: get head and use last-modified: Thu, 08 May 2025 08:35:35 GMT (curl --head https://nodejs.org/api/fs.html)
	if rc > 0 {
		fmt.Println("  AWESOME: found an individual project outside GitHub. saving it to the index")
		metadata = NewAwesomeLink()
		metadata.description = "Unknown, outside GitHub !"
		metadata.subscribers = 0
		metadata.watchers = 0
		metadata.clone_url = ""
		metadata.index_url = ""
		metadata.origin_url = repo
		fmt.Println("  SAVING: ", metadata)
		//saveLink(metadata)
		return
	}

	headers := make(map[string]string)
	headers["Accept"] = "application/vnd.github.v3.raw"

	rc, content := throttled_fetch(metadata.index_url, http.MethodGet, headers)

	if rc != 200 {
		log.Fatal("ERROR: ", model.RC_LINK_HAS_NO_INDEX_PAGE, ": ", rc)
	}

	if metadata.origin_url != model.AW_ROOT &&
		!is_awl_index_page(string(content)) {

		// we have reached an leaf, go get the project metadata before closing this branch
		fmt.Println("  AWESOME: found an individual project inside GitHub. saving it to the index")
		fmt.Println("  SAVING: ", metadata)
		// await saveLink(projectMeta);

		// terminating the branch
		return
	}

	fmt.Println("  looks like an Awesome index, walk over it.")

	content = remove_html_prefix(content)
	parse_markdown_index(content)

	// fmt.Printf("result is %s\n", result[0:31])
	fmt.Printf("rc is     %d\n", rc)
	fmt.Print(metadata)

	// 	console.debug(`  no index page found (${error.message})`);
	// return;
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

// walk up the tree and get all textual content in an ordered list
func get_parents(node ast.Node) {
	if node == nil {
		return
	}
	get_parents(node.GetParent())

	content := getContent(node)

	if len(content) > 0 {
		fmt.Println("PARENT: ", content)
	}

	return

	// if parent != nil {
	// 	fmt.Println("PARENT: ", getContent(parent))
	// }
}

func parse_markdown_index(md []byte) {

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
			fmt.Println("PARAGRAPH1: ", get_first_label_below_paragraph(paragraph))
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

// if there is a sibling type of "text" => use it
// if there is a sibling type of "link" => use it
// otherwise, placeholder with
func get_first_label_below_paragraph(node ast.Node) string {
	children := node.GetChildren()
	content := ""
	for _, node := range children {
		if child, ok := node.(*ast.Text); ok {
			content = getContent(child)
			break
		}
	}
	return content
}

// returne true/false if the given string (can be yaml) contains
// the Awesoem List content badge tag (and thus, is an AWL index)
func is_awl_index_page(content string) bool {
	return strings.Contains(content, "https://awesome.re/badge/")
}

// remove everything (html...) up to the first yaml heading tag ('^#')
func remove_html_prefix(content []byte) (result []byte) {
	pattern := regexp.MustCompile(`(?m)^#`)
	loc := pattern.FindIndex(content)

	result = content[loc[0]:]
	return
}
