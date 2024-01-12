/*
Copyright 2024 F. Hoffmann-La Roche AG

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
package cmd

import (
	"os"
	"strings"

	git "github.com/go-git/go-git/v5"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

func MirrorRepository(source, destination, sourcePatEnvVar, destinationPatEnvVar string) {
	gitDirectory, err := os.MkdirTemp(localTempDirectory, "")
	checkError(err)
	defer os.RemoveAll(gitDirectory)
	gitCloneOptions := &git.CloneOptions{
		URL: source,
		Auth: &githttp.BasicAuth{
			Username: "This can be any string.",
			Password: os.Getenv(sourcePatEnvVar)},
		Depth: 1,
	}
	repository, err := git.PlainClone(gitDirectory, false, gitCloneOptions)
	if err != nil {
		log.Error("Error while cloning ", source, ": ", err)
	}
	// Get SHA of repository HEAD.
	ref, err := repository.Head()
	checkError(err)
	log.Info(ref.Hash().String())
	remote, err := repository.Remote("origin")
	refList, err := remote.List(&git.ListOptions{})
	checkError(err)
	refPrefix := "refs/heads/"
	var branchList []string
	for _, ref := range refList {
		refName := ref.Name().String()
		if !strings.HasPrefix(refName, refPrefix) {
			continue
		}
		branchName := strings.TrimPrefix(refName, refPrefix)
		branchList = append(branchList, branchName)
	}
	log.Info("branches = ", branchList)
}
