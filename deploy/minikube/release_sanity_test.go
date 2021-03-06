/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"runtime"
	"testing"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"k8s.io/minikube/pkg/minikube/notify"
	"k8s.io/minikube/pkg/util"
)

func getSHAFromURL(url string) (string, error) {
	fmt.Println("Downloading: ", url)
	r, err := retryablehttp.Get(url)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	b := sha256.Sum256(body)
	return hex.EncodeToString(b[:]), nil
}

// TestReleasesJSON checks if all *GA* releases
// 	enlisted in https://storage.googleapis.com/minikube/releases.json
//	are available to download and have correct hashsum
func TestReleasesJSON(t *testing.T) {
	releases, err := notify.AllVersionsFromURL(notify.GithubMinikubeReleasesURL)
	if err != nil {
		t.Fatalf("Error getting releases.json: %v", err)
	}
	checkReleases(t, releases)
}

// TestBetaReleasesJSON checks if all *BETA* releases
// 	enlisted in https://storage.googleapis.com/minikube/releases-beta.json
//	are available to download and have correct hashsum
func TestBetaReleasesJSON(t *testing.T) {
	releases, err := notify.AllVersionsFromURL(notify.GithubMinikubeBetaReleasesURL)
	if err != nil {
		t.Fatalf("Error getting releases-bets.json: %v", err)
	}
	checkReleases(t, releases)
}

func checkReleases(t *testing.T, rs notify.Releases) {
	for _, r := range rs {
		fmt.Printf("Checking release: %s\n", r.Name)
		for platform, sha := range r.Checksums {
			fmt.Printf("Checking SHA for %s.\n", platform)
			actualSha, err := getSHAFromURL(util.GetBinaryDownloadURL(r.Name, platform, runtime.GOARCH))
			if err != nil {
				t.Errorf("Error calculating SHA for %s-%s. Error: %v", r.Name, platform, err)
				continue
			}
			if actualSha != sha {
				t.Errorf("ERROR: SHA does not match for version %s, platform %s. Expected %s, got %s.", r.Name, platform, sha, actualSha)
				continue
			}
		}
	}
}
