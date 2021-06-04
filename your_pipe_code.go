package main

//import "fmt"
import "regexp"
import "encoding/json"
//import "strings"

/*

   stage 2
   Where your filter code runs. The doc object is the h map

*/

// yourPipeCode objective is for a user to code the pipe transform stage in golang.
// That code would exist here and a ref to that completed map goes back in channel.
//
func yourPipeCode(h map[string]interface{}, c chan *map[string]interface{}) {

	// h is the hash representing your docs (which are a collection of fields)
	// keys are fieldnames
	// value is interface{} and must be asserted

	// Access log.file.path value
	// Legacy field
	//
	p := dotField(h, "log.file.path").(string)


	if "/var/log/gitlab/gitaly/current" == p {

		json.Unmarshal([]byte(h["message"].(string)), &h)
		addTags(&h, []string{"DecodedJsonToFields"})

		_, ok := h["grpc.code"]  // key exists
		isClone := regexp.MustCompile(`(SSH|Post)UploadPack`)
		isPush := regexp.MustCompile(`(SSH|InfoRef)ReceivePack`)
		isWikiRepo := regexp.MustCompile(`.wiki.git$`)

		// clone 
		if isClone.MatchString(h["grpc.method"].(string)) && ok && h["grpc.code"].(string) == "OK" {
			h["repo"] = h["grpc.request.glProjectPath"].(string)
			oniguruma("owner", "repo", &h, "([a-zA-Z0-9_-]+)(?:/*)")
			h["git_ops"] = "clone"
			h["ops_duration"] = h["grpc.time_ms"].(float64)

			// push 
		} else if isPush.MatchString(h["grpc.method"].(string)) && ok && h["grpc.code"].(string) == "OK" {
			h["repo"] = h["grpc.request.glProjectPath"].(string)
			oniguruma("owner", "repo", &h, "([a-zA-Z0-9_-]+)(?:/*)")
			h["git_ops"] = "push"
			h["ops_duration"] = h["grpc.time_ms"].(float64)

			// create
		} else if h["grpc.method"] == "CreateRepository" && !isWikiRepo.MatchString(h["grpc.request.repoPath"].(string)) &&
			ok && h["grpc.code"].(string) == "OK" {
			h["repo"] = h["grpc.request.glProjectPath"].(string)
			oniguruma("owner", "repo", &h, "([a-zA-Z0-9_-]+)(?:/*)")
			h["git_ops"] = "create"

			// delete
		} else if h["grpc.method"] == "RemoveRepository" && !isWikiRepo.MatchString(h["grpc.request.repoPath"].(string)) &&
			ok && h["grpc.code"].(string) == "OK" {
			h["git_ops"] = "delete"
		}

	}

	c <- &h // Although you write code here this line is required
}
