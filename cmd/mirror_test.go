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
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SetRepositoryAuth(t *testing.T) {
	repositories := []RepositoryPair{
		{
			Repository{
				"https://example.com/org-1/repo-1",
				Authentication{"", ""},
			},
			Repository{
				"https://example.com/org-2/repo-2",
				Authentication{"", ""},
			},
		},
		{
			Repository{
				"https://example.com/org-3/repo-3",
				Authentication{"token", "CUSTOM_TOKEN_1"},
			},
			Repository{
				"https://example.com/org-4/repo-4",
				Authentication{"token", "CUSTOM_TOKEN_2"},
			},
		},
	}
	defaultSettings := RepositoryPair{
		Repository{
			"", Authentication{"token", "GITLAB_TOKEN"},
		},
		Repository{
			"", Authentication{"token", "GITHUB_TOKEN"},
		},
	}
	SetRepositoryAuth(&repositories, defaultSettings)
	assert.Equal(t, repositories[0].Source.Auth.Method, "token")
	assert.Equal(t, repositories[0].Source.Auth.TokenName, "GITLAB_TOKEN")
	assert.Equal(t, repositories[0].Destination.Auth.Method, "token")
	assert.Equal(t, repositories[0].Destination.Auth.TokenName, "GITHUB_TOKEN")
	assert.Equal(t, repositories[1].Source.Auth.Method, "token")
	assert.Equal(t, repositories[1].Source.Auth.TokenName, "CUSTOM_TOKEN_1")
	assert.Equal(t, repositories[1].Destination.Auth.Method, "token")
	assert.Equal(t, repositories[1].Destination.Auth.TokenName, "CUSTOM_TOKEN_2")
}
