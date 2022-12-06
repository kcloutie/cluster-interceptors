package pkg

import triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"

type payloadDetails struct {
	PrNumber     int
	Owner        string
	Repository   string
	ChangedFiles changedFiles
}

type changedFiles struct {
	Files    string
	FileList []string
}

// GitHubInterceptor provides a webhook to add changed files to a pull request event
type GitHubAddChangeInterceptor struct {
	SecretRef *triggersv1.SecretRef `json:"secretRef,omitempty"`
}
