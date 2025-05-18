package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const GH_API_URL = "https://api.github.com"
const AW_ROOT = "https://github.com/sindresorhus/awesome"

const RC_OK = 0x0                         // success
const RC_LINK_HAS_NO_INDEX_PAGE = 0x1     // the provided has no index page (currently only readme type is implemented)
const RC_LINK_IS_NOT_A_PROJECT_ROOT = 0x2 // link is towards a either a user landing page or a subfolder someware under the project root
const RC_LINK_IS_NOT_ON_GITHUB = 0x3

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
		rc = RC_LINK_IS_NOT_ON_GITHUB
		return
	}

	url_parts := strings.Split(url, "/")
	if len(url_parts) > 5 {
		rc = RC_LINK_IS_NOT_A_PROJECT_ROOT
		return
	}

	username := url_parts[3]
	repo := url_parts[4]
	headers := make(map[string]string)

	to_try := [2]string{"README.md", "readme.md"}
	for _, page := range to_try {
		fmt.Print("trying: ", page, " ... ")
		target := fmt.Sprintf("%s/repos/%s/%s/contents/%s", GH_API_URL, username, repo, page)
		rc, _ := throttled_fetch(target, http.MethodHead, headers)
		fmt.Println(rc)
		if rc == 404 {
			continue
		}
		metadata.index_url = target + "/" + page
	}

	if metadata.index_url == "" {
		rc = RC_LINK_HAS_NO_INDEX_PAGE
		return
	}

	target := fmt.Sprintf("%s/repos/%s/%s", GH_API_URL, username, repo)
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

func index(repo string, depth int) {
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

	rc, res := throttled_fetch(metadata.index_url, http.MethodGet, headers)

	fmt.Print(res)

	// fmt.Printf("result is %s\n", result[0:31])
	fmt.Printf("rc is     %d\n", rc)
	fmt.Print(metadata)

	// 	console.debug(`  no index page found (${error.message})`);
	// return;
}

func main() {
	_ = godotenv.Load()

	index(AW_ROOT, 0)
	// fmt.Print(json)
}
