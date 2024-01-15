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
	"github.com/go-git/go-git/v5/config"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

const refBranchPrefix = "refs/heads/"
const refTagPrefix = "refs/tags/"
const basicAuthUsername = "This can be any string."

func GetBranchesAndTagsFromRemote(repository *git.Repository, remoteName string, listOptions *git.ListOptions) ([]string, []string) {
	remote, err := repository.Remote(remoteName)
	checkError(err)
	refList, err := remote.List(listOptions)
	checkError(err)
	var branchList []string
	var tagList []string
	for _, ref := range refList {
		refName := ref.Name().String()
		if strings.HasPrefix(refName, refBranchPrefix) {
			branchName := strings.TrimPrefix(refName, refBranchPrefix)
			branchList = append(branchList, branchName)
		} else if strings.HasPrefix(refName, refTagPrefix) {
			tagName := strings.TrimPrefix(refName, refTagPrefix)
			tagList = append(tagList, tagName)
		}
	}
	return branchList, tagList
}

func ProcessPushingError(err error, url string, activity string, allErrors *[]string) {
	var e string
	if err != nil && err != git.NoErrAlreadyUpToDate {
		e = "Error while " + activity + url + ": " + err.Error()
	}
	if e != "" {
		*allErrors = append(*allErrors, e)
	}
}

func MirrorRepository(source, destination, sourcePatEnvVar, destinationPatEnvVar string) {
	log.Debug("Cloning ", source, "...")
	gitDirectory, err := os.MkdirTemp(localTempDirectory, "")
	checkError(err)
	defer os.RemoveAll(gitDirectory)
	sourceAuth := &githttp.BasicAuth{
		Username: basicAuthUsername,
		Password: os.Getenv(sourcePatEnvVar)}
	gitCloneOptions := &git.CloneOptions{
		URL:  source,
		Auth: sourceAuth,
	}
	repository, err := git.PlainClone(gitDirectory, false, gitCloneOptions)
	if err != nil {
		log.Error("Error while cloning ", source, ": ", err)
	}
	sourceBranchList, sourceTagList := GetBranchesAndTagsFromRemote(repository, "origin", &git.ListOptions{})
	log.Debug(source, " branches = ", sourceBranchList)
	log.Debug(source, " tags = ", sourceTagList)

	log.Info("Fetching all branches from ", source, "...")
	sourceRemote, err := repository.Remote("origin")
	checkError(err)
	sourceRemote.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{"refs/heads/*:refs/heads/*"},
		Auth:     sourceAuth,
	})

	_, err = repository.CreateRemote(&config.RemoteConfig{
		Name: "destination",
		URLs: []string{destination},
	})
	checkError(err)

	destinationAuth := &githttp.BasicAuth{
		Username: basicAuthUsername,
		Password: os.Getenv(destinationPatEnvVar)}

	destinationBranchList, destinationTagList := GetBranchesAndTagsFromRemote(repository, "destination", &git.ListOptions{Auth: destinationAuth})
	log.Debug(destination, " branches = ", destinationBranchList)
	log.Debug(destination, " tags = ", destinationTagList)

	var allErrors []string

	log.Info("Pushing all branches from ", source, " to ", destination)
	for _, branch := range sourceBranchList {
		log.Debug("Pushing branch ", branch)
		err := repository.Push(&git.PushOptions{
			RemoteName: "destination",
			RefSpecs:   []config.RefSpec{config.RefSpec("+" + refBranchPrefix + branch + ":" + refBranchPrefix + branch)},
			Auth:       destinationAuth, Force: true, Atomic: true})
		ProcessPushingError(err, destination, "pushing branch "+branch+" to ", &allErrors)

	}
	// Remove any branches not present in the source repository anymore.
	for _, branch := range destinationBranchList {
		if !stringInSlice(branch, sourceBranchList) {
			log.Info("Removing branch ", branch, " from ", destination)
			err := repository.Push(&git.PushOptions{
				RemoteName: "destination",
				RefSpecs:   []config.RefSpec{config.RefSpec(":" + refBranchPrefix + branch)},
				Auth:       destinationAuth, Force: true, Atomic: true})
			ProcessPushingError(err, destination, "removing branch "+branch+" from ", &allErrors)
		}
	}
	log.Info("Pushing all tags from ", source, " to ", destination)
	for _, tag := range sourceTagList {
		log.Debug("Pushing tag ", tag)
		err := repository.Push(&git.PushOptions{
			RemoteName: "destination",
			RefSpecs:   []config.RefSpec{config.RefSpec("+" + refTagPrefix + tag + ":" + refTagPrefix + tag)},
			Auth:       destinationAuth, Force: true, Atomic: true})
		ProcessPushingError(err, destination, "pushing tag "+tag+" to ", &allErrors)
	}
	// Remove any tags not present in the source repository anymore.
	for _, tag := range destinationTagList {
		if !stringInSlice(tag, sourceTagList) {
			log.Info("Removing tag ", tag, " from ", destination)
			err := repository.Push(&git.PushOptions{
				RemoteName: "destination",
				RefSpecs:   []config.RefSpec{config.RefSpec(":" + refTagPrefix + tag)},
				Auth:       destinationAuth, Force: true, Atomic: true})
			ProcessPushingError(err, destination, "removing tag "+tag+" from ", &allErrors)
		}
	}
	for _, e := range allErrors {
		log.Error(e)
	}
}
