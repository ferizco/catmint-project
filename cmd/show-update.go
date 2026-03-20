package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	githubOwner = "ferizco"
	githubRepo  = "catmint-project"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

func runShowUpdate(_ []string) {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/releases/latest",
		githubOwner,
		githubRepo,
	)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("Unable to check updates (offline?)")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "GitHub API error: %s\n", resp.Status)
		os.Exit(1)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	current := version
	latest := release.TagName

	fmt.Printf("Current version : %s\n", version)
	fmt.Printf("Latest version  : %s\n", release.TagName)

	if current == latest {
		fmt.Println("You are using the latest version 👍")
	} else {
		fmt.Println("Update available 🚀")
		fmt.Printf("See: %s\n", release.HTMLURL)
	}
}
