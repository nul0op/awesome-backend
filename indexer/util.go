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
		rc, _ := fetch(target, http.MethodHead, headers)
		fmt.Println(rc)
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
	fmt.Println(strings.Join(topics, ":"))

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

// walk up the tree and get all textual content in an ordered list
func getParentsPath(node ast.Node) (parents []string) {
	parents = []string{}

	if node == nil {
		return
	}

	getParentsPath(node.GetParent())

	content := getContent(node)
	if len(content) > 0 {
		fmt.Println("PARENT: ", content)
	}

	return
}

// if there is a sibling type of "text" => use it
// if there is a sibling type of "link" => use it
// otherwise, placeholder with
func getFirstChildLabel(node ast.Node) string {
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
	log.Println("DEBUG: fetching url: ", url)

	// FIXME: look at the header and throttle when the remote warn us instead
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
