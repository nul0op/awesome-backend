package indexer

import (
	"awesome-portal/backend/model"
	"awesome-portal/backend/util"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func stripHtmlAnchor(content string) (result string) {
	result = content
	pattern := regexp.MustCompile(`#.*$`)
	loc := pattern.FindIndex([]byte(content))

	if len(loc) > 0 {
		result = content[0:loc[0]]
	}

	return
}

func getProjectMetaData(url string) (metadata model.AwesomeLink, rc int) {

	url = stripHtmlAnchor(url)

	if strings.Index(url, "https://github.com/") != 0 {
		rc = model.RC_LINK_IS_NOT_ON_GITHUB
		return
	}

	url_parts := strings.Split(url, "/")
	if len(url_parts) > 5 {
		rc = model.RC_LINK_IS_NOT_A_PROJECT_ROOT
		return
	} else if len(url_parts) < 5 {
		rc = model.RC_LINK_IS_A_USER_LANDING_PAGE
		return
	}

	// at this point, we can start to investigate and get some usefull data
	metadata = model.NewAwesomeLink()

	username := url_parts[3]
	repo := url_parts[4]
	headers := make(map[string]string)

	to_try := [4]string{"README.md", "readme.md", "README.textile", "readme.textile"}
	for _, page := range to_try {
		target := fmt.Sprintf("%s/repos/%s/%s/contents/%s", model.GH_API_URL, username, repo, page)
		rc, _ := callGithubAPI(target, http.MethodHead, headers)
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
	rc, payload := callGithubAPI(target, http.MethodGet, headers)

	// 	// err := json.NewDecoder(data).Decode(&gh_response)

	kv := payloadToJson(payload)

	metadata.Name = kv["name"].(string)
	if kv["description"] != nil {
		metadata.Description = kv["description"].(string)
	}
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
		topics = append(topics, v.(string))
	}
	// util.Log.Debug(strings.Join(topics, ":"))

	rc = 0
	return
}

// ---------------------------- util for REST api calls

func payloadToJson(data []byte) map[string]interface{} {
	kv := make(map[string]interface{})

	err := json.Unmarshal(data, &kv)
	if err != nil {
		util.Log.Errorf("%w\n", err)
	}

	// for key, value := range kv {
	// 	fmt.Println(key, value)
	// }

	return kv
}

func callGithubAPI(url string, method string, headers map[string]string) (rc int, payload []byte) {
	// util.Log.Debugf("fetching: %s", url)

	// FIXME: look at the header and throttle when the remote warn us instead
	time.Sleep(250 * time.Millisecond)

	gh_client := http.Client{
		Timeout: time.Second * 5,
	}

	req, error := http.NewRequest(method, url, nil)
	if error != nil {
		util.Log.Error(error)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("GH_TOKEN")))
	for name, value := range headers {
		req.Header.Add(name, value)
	}

	res, error := gh_client.Do(req)
	if error != nil {
		rc = 1
		util.Log.Error(error)
		return
	}
	rc = res.StatusCode

	if res.Header["X-Ratelimit-Remaining"] != nil {
		reqLeft, _ := strconv.Atoi(res.Header.Get("X-Ratelimit-Remaining"))
		if reqLeft < 1000 {
			util.Log.Warnf("not enough request left, throthling: %d", reqLeft)
			time.Sleep(5 * time.Second)
		} else if reqLeft%10 == 0 {
			util.Log.Warnf("requests left: %d", reqLeft)
		}
	}

	payload, error = io.ReadAll(res.Body)

	if error != nil {
		util.Log.Error(error)
	}

	return
}
