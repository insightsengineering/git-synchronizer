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
	"encoding/json"
	"os"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

const refBranchPrefix = "refs/heads/"
const refTagPrefix = "refs/tags/"
const basicAuthUsername = "This can be any string."
const token = "token"

type MirrorStatus struct {
	Errors        []string
	LastCloneEnd  time.Time
	CloneDuration time.Duration
	PushDuration  time.Duration
}

// SetRepositoryAuth ensures that repositories for which the authentication settings have not been
// overridden, use the default authentication settings from config file.
func SetRepositoryAuth(repositories *[]RepositoryPair, defaultSettings RepositoryPair) {
	for i := 0; i < len(*repositories); i++ {
		if (*repositories)[i].Source.Auth.Method == "" {
			(*repositories)[i].Source.Auth.Method = defaultSettings.Source.Auth.Method
			if (*repositories)[i].Source.Auth.Method == token {
				(*repositories)[i].Source.Auth.TokenName = defaultSettings.Source.Auth.TokenName
			}
		}
		if (*repositories)[i].Destination.Auth.Method == "" {
			(*repositories)[i].Destination.Auth.Method = defaultSettings.Destination.Auth.Method
			if (*repositories)[i].Destination.Auth.Method == token {
				(*repositories)[i].Destination.Auth.TokenName = defaultSettings.Destination.Auth.TokenName
			}
		}
	}
	repositoriesJSON, err := json.MarshalIndent(*repositories, "", "  ")
	checkError(err)
	log.Trace("repositories = ", string(repositoriesJSON))
}

// ValidateRepositories checks for common issues with input repository data from config file.
func ValidateRepositories(repositories []RepositoryPair) {
	var allDestinationRepositories []string
	for _, repo := range repositories {
		if stringInSlice(repo.Destination.RepositoryURL, allDestinationRepositories) {
			log.Fatal(
				"Error: multiple repositories set to be synchronized to the same destination repository: ",
				repo.Source.RepositoryURL,
			)
		}
		allDestinationRepositories = append(allDestinationRepositories, repo.Destination.RepositoryURL)
		sourceURL := strings.Split(repo.Source.RepositoryURL, "/")
		sourceProjectName := sourceURL[len(sourceURL)-1]
		destinationURL := strings.Split(repo.Destination.RepositoryURL, "/")
		destinationProjectName := destinationURL[len(destinationURL)-1]
		if sourceProjectName != destinationProjectName {
			log.Warn(
				"Warning: Source project name (", sourceProjectName,
				") and destination project name (", destinationProjectName, ") differ!",
			)
		}
	}
}

// GetBranchesAndTagsFromRemote returns list of branches and tags present in remoteName of repository.
func GetBranchesAndTagsFromRemote(repository *git.Repository, remoteName string, listOptions *git.ListOptions) ([]string, []string, error) {
	var branchList []string
	var tagList []string
	remote, err := repository.Remote(remoteName)
	if err != nil {
		return branchList, tagList, err
	}
	refList, err := remote.List(listOptions)
	if err != nil {
		return branchList, tagList, err
	}

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
	return branchList, tagList, nil
}

// ProcessError formats err and appends it to allErrors.
func ProcessError(err error, activity string, url string, allErrors *[]string) {
	var e string
	if err != nil && err != git.NoErrAlreadyUpToDate {
		e = "Error while " + activity + url + ": " + err.Error()
	}
	if e != "" {
		log.Error(e)
		*allErrors = append(*allErrors, e)
	}
}

// GetCloneOptions returns clone options for source repository.
func GetCloneOptions(source string, sourceAuth Authentication) *git.CloneOptions {
	var sourcePat string
	if sourceAuth.Method == token {
		sourcePat = os.Getenv(sourceAuth.TokenName)
	} else if sourceAuth.Method != "" {
		log.Error("Unknown auth method: ", sourceAuth.Method)
	}
	if sourcePat != "" {
		gitCloneOptions := &git.CloneOptions{
			URL: source,
			Auth: &githttp.BasicAuth{
				Username: basicAuthUsername,
				Password: sourcePat,
			},
		}
		return gitCloneOptions
	}
	gitCloneOptions := &git.CloneOptions{URL: source}
	return gitCloneOptions
}

// GetListOptions returns list options for source repository.
func GetListOptions(sourceAuth Authentication) *git.ListOptions {
	var sourcePat string
	if sourceAuth.Method == token {
		sourcePat = os.Getenv(sourceAuth.TokenName)
	} else if sourceAuth.Method != "" {
		log.Error("Unknown auth method: ", sourceAuth.Method)
	}
	if sourcePat != "" {
		gitListOptions := &git.ListOptions{
			Auth: &githttp.BasicAuth{
				Username: basicAuthUsername,
				Password: sourcePat,
			},
		}
		return gitListOptions
	}
	gitListOptions := &git.ListOptions{}
	return gitListOptions
}

// GetFetchOptions returns fetch options for source repository.
func GetFetchOptions(refSpec string, sourceAuth Authentication) *git.FetchOptions {
	var sourcePat string
	if sourceAuth.Method == token {
		sourcePat = os.Getenv(sourceAuth.TokenName)
	} else if sourceAuth.Method != "" {
		log.Error("Unknown auth method: ", sourceAuth.Method)
	}
	if sourcePat != "" {
		gitFetchOptions := &git.FetchOptions{
			RefSpecs: []config.RefSpec{config.RefSpec(refSpec)},
			Auth: &githttp.BasicAuth{
				Username: basicAuthUsername,
				Password: sourcePat,
			},
		}
		return gitFetchOptions
	}
	gitFetchOptions := &git.FetchOptions{
		RefSpecs: []config.RefSpec{config.RefSpec(refSpec)},
	}
	return gitFetchOptions
}

// GetDestionationAuth returns authentication struct for destination git repository.
func GetDestinationAuth(destAuth Authentication) *githttp.BasicAuth {
	var destinationPat string
	if destAuth.Method == token {
		destinationPat = os.Getenv(destAuth.TokenName)
	} else if destAuth.Method != "" {
		log.Error("Unknown auth method: ", destAuth.Method)
	}
	destinationAuth := &githttp.BasicAuth{
		Username: basicAuthUsername,
		Password: destinationPat,
	}
	return destinationAuth
}

// MirrorRepository mirrors branches and tags from source to destination. Tags and branches
// no longer present in source are removed from destination.
func MirrorRepository(messages chan MirrorStatus, source, destination string, sourceAuthentication, destinationAuthentication Authentication) {
	log.Debug("Cloning ", source)
	cloneStart := time.Now()
	gitDirectory, err := os.MkdirTemp(localTempDirectory, "")
	checkError(err)
	defer os.RemoveAll(gitDirectory)
	var allErrors []string
	gitCloneOptions := GetCloneOptions(source, sourceAuthentication)
	repository, err := git.PlainClone(gitDirectory, false, gitCloneOptions)
	if err != nil {
		ProcessError(err, "cloning repository from ", source, &allErrors)
		messages <- MirrorStatus{allErrors, time.Now(), 0, 0}
		return
	}

	gitListOptions := GetListOptions(sourceAuthentication)
	sourceBranchList, sourceTagList, err := GetBranchesAndTagsFromRemote(repository, "origin", gitListOptions)
	if err != nil {
		ProcessError(err, "getting branches and tags from ", source, &allErrors)
		messages <- MirrorStatus{allErrors, time.Now(), 0, 0}
		return
	}
	log.Debug(source, " branches = ", sourceBranchList)
	log.Debug(source, " tags = ", sourceTagList)

	log.Info("Fetching all branches from ", source)
	sourceRemote, err := repository.Remote("origin")
	if err != nil {
		ProcessError(err, "getting source remote for ", source, &allErrors)
		messages <- MirrorStatus{allErrors, time.Now(), 0, 0}
		return
	}

	gitFetchOptions := GetFetchOptions("refs/heads/*:refs/heads/*", sourceAuthentication)
	err = sourceRemote.Fetch(gitFetchOptions)
	if err != nil {
		ProcessError(err, "fetching branches from ", source, &allErrors)
		messages <- MirrorStatus{allErrors, time.Now(), 0, 0}
		return
	}

	cloneDuration := time.Since(cloneStart)
	cloneEnd, pushStart := time.Now(), time.Now()

	_, err = repository.CreateRemote(&config.RemoteConfig{
		Name: "destination",
		URLs: []string{destination},
	})
	if err != nil {
		ProcessError(err, "creating remote for ", destination, &allErrors)
		messages <- MirrorStatus{allErrors, time.Now(), 0, 0}
		return
	}

	destinationAuth := GetDestinationAuth(destinationAuthentication)

	destinationBranchList, destinationTagList, err := GetBranchesAndTagsFromRemote(repository, "destination", &git.ListOptions{Auth: destinationAuth})
	if err != nil {
		ProcessError(err, "getting branches and tags from ", destination, &allErrors)
		messages <- MirrorStatus{allErrors, time.Now(), 0, 0}
		return
	}
	log.Debug(destination, " branches = ", destinationBranchList)
	log.Debug(destination, " tags = ", destinationTagList)

	log.Info("Pushing all branches from ", source, " to ", destination)
	for _, branch := range sourceBranchList {
		log.Debug("Pushing branch ", branch, " to ", destination)
		err = repository.Push(&git.PushOptions{
			RemoteName: "destination",
			RefSpecs:   []config.RefSpec{config.RefSpec("+" + refBranchPrefix + branch + ":" + refBranchPrefix + branch)},
			Auth:       destinationAuth, Force: true, Atomic: true})
		ProcessError(err, "pushing branch "+branch+" to ", destination, &allErrors)
	}

	// Remove any branches not present in the source repository anymore.
	for _, branch := range destinationBranchList {
		if !stringInSlice(branch, sourceBranchList) {
			log.Info("Removing branch ", branch, " from ", destination)
			err = repository.Push(&git.PushOptions{
				RemoteName: "destination",
				RefSpecs:   []config.RefSpec{config.RefSpec(":" + refBranchPrefix + branch)},
				Auth:       destinationAuth, Force: true, Atomic: true})
			ProcessError(err, "removing branch "+branch+" from ", destination, &allErrors)
		}
	}

	log.Info("Pushing all tags from ", source, " to ", destination)
	err = repository.Push(&git.PushOptions{
		RemoteName: "destination",
		RefSpecs:   []config.RefSpec{config.RefSpec("+" + refTagPrefix + "*:" + refTagPrefix + "*")},
		Auth:       destinationAuth, Force: true, Atomic: true})
	ProcessError(err, "pushing all tags to ", destination, &allErrors)

	// Remove any tags not present in the source repository anymore.
	for _, tag := range destinationTagList {
		if !stringInSlice(tag, sourceTagList) {
			log.Info("Removing tag ", tag, " from ", destination)
			err := repository.Push(&git.PushOptions{
				RemoteName: "destination",
				RefSpecs:   []config.RefSpec{config.RefSpec(":" + refTagPrefix + tag)},
				Auth:       destinationAuth, Force: true, Atomic: true})
			ProcessError(err, "removing tag "+tag+" from ", destination, &allErrors)
		}
	}
	pushDuration := time.Since(pushStart)
	messages <- MirrorStatus{allErrors, cloneEnd, cloneDuration, pushDuration}
}

// MirrorRepositories ensures that branches and tags from source repository are mirrored to
// the destination repository for each repositoryPair.
func MirrorRepositories(repos []RepositoryPair) {
	messages := make(chan MirrorStatus, 100)
	var allErrors []string
	synchronizationStart := time.Now()
	for _, repository := range repos {
		log.Info("Mirroring ", repository.Source.RepositoryURL, " → ", repository.Destination.RepositoryURL)
		go MirrorRepository(
			messages, repository.Source.RepositoryURL, repository.Destination.RepositoryURL,
			repository.Source.Auth, repository.Destination.Auth,
		)
	}
	receivedResults := 0
	var lastCloneEnd time.Time
	var totalCloneDuration time.Duration
	var totalPushDuration time.Duration
results_receiver_loop:
	for {
		select {
		case msg := <-messages:
			receivedResults++
			log.Info("Finished mirroring ", receivedResults, " out of ", len(repos), " repositories.")
			combineSlices(msg.Errors, &allErrors)
			if lastCloneEnd.Before(msg.LastCloneEnd) {
				lastCloneEnd = msg.LastCloneEnd
			}
			totalCloneDuration += msg.CloneDuration
			totalPushDuration += msg.PushDuration
			if receivedResults == len(repos) {
				break results_receiver_loop
			}
		default:
			time.Sleep(time.Second)
		}
	}
	cloneDuration := lastCloneEnd.Sub(synchronizationStart)
	syncDuration := time.Since(synchronizationStart)
	log.Infof("Last clone finished %v after synchronization had started (%.1f%% of total synchronization time).",
		cloneDuration.Round(time.Second), (float64(100)*cloneDuration.Seconds())/syncDuration.Seconds())
	log.Infof("Synchronization took %v (wall-clock time).", syncDuration.Round(time.Second))
	log.Debugf("Total clone duration: %v (goroutine time).", totalCloneDuration.Round(time.Second))
	log.Debugf("Total push duration: %v (goroutine time).", totalPushDuration.Round(time.Second))
	if len(allErrors) > 0 {
		log.Error("The following errors have been encountered:")
		for _, e := range allErrors {
			log.Error(e)
		}
		os.Exit(1)
	}
}
