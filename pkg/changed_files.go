/*
 Copyright 2022 The Tekton Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	gh "github.com/google/go-github/v31/github"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
	"github.com/tektoncd/triggers/pkg/interceptors"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
)

const (
	changedFilesExtensionsKey = "changed_files"
)

var _ triggersv1.InterceptorInterface = (*Interceptor)(nil)

// ErrInvalidContentType is returned when the content-type is not a JSON body.
var ErrInvalidContentType = errors.New("form parameter encoding not supported, please change the hook to send JSON payloads")
var acceptedEventTypes = []string{"pull_request", "merge"}

type Interceptor struct {
	// AuthToken is an OAuth token used to connect to the GitHub API
	// AuthToken string
	SecretGetter interceptors.SecretGetter
}

func NewInterceptor(sg interceptors.SecretGetter) *Interceptor {
	return &Interceptor{
		SecretGetter: sg,
	}
}

// Process(ctx context.Context, r *           InterceptorRequest) *           InterceptorResponse
func (w Interceptor) Process(ctx context.Context, r *triggersv1.InterceptorRequest) *triggersv1.InterceptorResponse {
	headers := interceptors.Canonical(r.Header)
	if v := headers.Get("Content-Type"); v == "application/x-www-form-urlencoded" {
		return interceptors.Fail(codes.InvalidArgument, ErrInvalidContentType.Error())
	}

	p := GitHubAddChangeInterceptor{}
	if err := interceptors.UnmarshalParams(r.InterceptorParams, &p); err != nil {
		return interceptors.Failf(codes.InvalidArgument, "failed to parse interceptor params: %v", err)
	}

	actualEvent := headers.Get("X-GitHub-Event")
	isAllowed := false
	for _, allowedEvent := range acceptedEventTypes {
		if actualEvent == allowedEvent {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		return interceptors.Failf(codes.FailedPrecondition, "event type %s is not allowed", actualEvent)
	}

	secretToken, err := w.getSecret(ctx, r, p)
	if err != nil {
		return interceptors.Failf(codes.FailedPrecondition, "error getting secret: %v", err)
	}

	payload, err := parseBody(r.Body, actualEvent)
	if err != nil {
		return interceptors.Failf(codes.FailedPrecondition, "error parsing body: %v", err)
	}

	var changedFiles changedFiles
	if actualEvent == "pull_request" {
		changedFiles, err = getChangedFilesFromPr(ctx, payload, p.EnterpriseBaseURL, secretToken)
		if err != nil {
			return interceptors.Failf(codes.FailedPrecondition, "error getting changed files: %v", err)
		}
	} else {
		changedFiles = payload.ChangedFiles
	}

	return &triggersv1.InterceptorResponse{
		Extensions: map[string]interface{}{
			changedFilesExtensionsKey: changedFiles,
		},
		Continue: true,
	}
}

func (w Interceptor) getSecret(ctx context.Context, r *triggersv1.InterceptorRequest, p GitHubAddChangeInterceptor) (string, error) {
	if p.SecretRef == nil {
		return "", nil
	}
	if p.SecretRef.SecretKey == "" {
		return "", fmt.Errorf("github interceptor secretRef.secretKey is empty")
	}
	ns, _ := triggersv1.ParseTriggerID(r.Context.TriggerID)
	secretToken, err := w.SecretGetter.Get(ctx, ns, p.SecretRef)
	if err != nil {
		return "", err
	}
	return string(secretToken), nil
}

func parseBody(body string, eventType string) (payloadDetails, error) {
	results := payloadDetails{}
	if body == "" {
		return results, fmt.Errorf("body was empty")
	}

	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(body), &jsonMap)
	if err != nil {
		return results, err
	}

	var prNum int
	_, ok := jsonMap["number"]
	if !ok && eventType == "pull_request" {
		return results, fmt.Errorf("pull_request body missing 'number' field")
	} else if eventType == "pull_request" {
		prNum = int(jsonMap["number"].(float64))
	} else {
		prNum = -1
	}

	repoSection, ok := jsonMap["repository"].(map[string]interface{})
	if !ok {
		return results, fmt.Errorf("payload body missing 'repository' field")
	}

	fullName, ok := repoSection["full_name"].(string)
	if !ok {
		return results, fmt.Errorf("payload body missing 'repository.full_name' field")
	}

	changedFiles := changedFiles{
		All:      []string{},
		Added:    []string{},
		Removed:  []string{},
		Modified: []string{},
	}

	commitsSection, ok := jsonMap["commits"].([]interface{})
	if ok {

		for _, commit := range commitsSection {
			addedFiles, ok := commit.(map[string]interface{})["added"].([]interface{})
			if !ok {
				return results, fmt.Errorf("payload body missing 'commits.*.added' field")
			}

			modifiedFiles, ok := commit.(map[string]interface{})["modified"].([]interface{})
			if !ok {
				return results, fmt.Errorf("payload body missing 'commits.*.modified' field")
			}

			removedFiles, ok := commit.(map[string]interface{})["removed"].([]interface{})
			if !ok {
				return results, fmt.Errorf("payload body missing 'commits.*.removed' field")
			}
			for _, fileName := range addedFiles {
				changedFiles.All = append(changedFiles.All, fmt.Sprintf("%v", fileName))
				changedFiles.Added = append(changedFiles.Added, fmt.Sprintf("%v", fileName))
			}

			for _, fileName := range modifiedFiles {
				changedFiles.All = append(changedFiles.All, fmt.Sprintf("%v", fileName))
				changedFiles.Modified = append(changedFiles.Modified, fmt.Sprintf("%v", fileName))
			}

			for _, fileName := range removedFiles {
				changedFiles.All = append(changedFiles.All, fmt.Sprintf("%v", fileName))
				changedFiles.Removed = append(changedFiles.Removed, fmt.Sprintf("%v", fileName))
			}
		}

		changedFiles.AllString = strings.Join(changedFiles.All, ",")
	}

	results = payloadDetails{
		PrNumber:     prNum,
		Owner:        strings.Split(fullName, "/")[0],
		Repository:   strings.Split(fullName, "/")[1],
		ChangedFiles: changedFiles,
	}
	return results, nil
}

func getChangedFilesFromPr(ctx context.Context, payload payloadDetails, enterpriseBaseURL string, token string) (changedFiles, error) {
	var httpClient *http.Client
	var client *gh.Client
	var err error
	changedFiles := changedFiles{
		All:      []string{},
		Added:    []string{},
		Removed:  []string{},
		Modified: []string{},
	}
	if token != "" {
		tokenSource := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		httpClient = oauth2.NewClient(ctx, tokenSource)
	} else {
		httpClient = nil
	}

	if enterpriseBaseURL != "" {
		client, err = gh.NewEnterpriseClient(enterpriseBaseURL, enterpriseBaseURL, httpClient)
		if err != nil {
			return changedFiles, err
		}
	} else {
		client = gh.NewClient(httpClient)
	}

	opt := &gh.ListOptions{PerPage: 100}
	// get all pages of results
	// var allCommitFiles []*gh.CommitFile
	for {
		files, resp, err := client.PullRequests.ListFiles(ctx, payload.Owner, payload.Repository, payload.PrNumber, opt)
		if err != nil {
			return changedFiles, err
		}
		for _, file := range files {
			changedFiles.All = append(changedFiles.All, *file.Filename)

			if *file.Status == "added" {
				changedFiles.Added = append(changedFiles.Added, *file.Filename)
			}
			if *file.Status == "modified" {
				changedFiles.Modified = append(changedFiles.Modified, *file.Filename)
			}
			if *file.Status == "removed" {
				changedFiles.Removed = append(changedFiles.Removed, *file.Filename)
			}

		}

		// allCommitFiles = append(allCommitFiles, files...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	changedFiles.AllString = strings.Join(changedFiles.All, ",")

	return changedFiles, nil
}
