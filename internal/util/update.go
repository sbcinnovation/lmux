package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// CheckForUpdate queries GitHub for the latest release tag and compares it with the
// provided currentVersion. It returns the latest version string (without leading 'v'),
// whether an update is available, the release URL if known, and an error if any.
func CheckForUpdate(ownerRepo string, currentVersion string) (latest string, updateAvailable bool, url string, err error) {
	api := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", ownerRepo)

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return "", false, "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "lmux-update-check")

	resp, err := client.Do(req)
	if err != nil {
		return "", false, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", false, "", errors.New("failed to fetch latest release")
	}

	var gr githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return "", false, "", err
	}

	latest = strings.TrimPrefix(strings.TrimSpace(gr.TagName), "v")
	current := strings.TrimPrefix(strings.TrimSpace(currentVersion), "v")
	if latest == "" || current == "" {
		return latest, false, gr.HTMLURL, nil
	}

	cmp := compareSemver(current, latest)
	return latest, cmp < 0, gr.HTMLURL, nil
}

// compareSemver compares semantic version strings (MAJOR.MINOR.PATCH[-prerelease]).
// Returns -1 if a < b, 0 if equal, 1 if a > b. Very small and permissive.
func compareSemver(a, b string) int {
	// Drop build metadata
	a = strings.SplitN(a, "+", 2)[0]
	b = strings.SplitN(b, "+", 2)[0]
	// Split prerelease
	aMain, aPre := splitPrerelease(a)
	bMain, bPre := splitPrerelease(b)

	// Compare core parts
	aParts := splitInts(aMain)
	bParts := splitInts(bMain)
	for i := 0; i < 3; i++ {
		if aParts[i] < bParts[i] {
			return -1
		}
		if aParts[i] > bParts[i] {
			return 1
		}
	}

	// If core equal, handle prerelease: absence > presence
	if aPre == "" && bPre == "" {
		return 0
	}
	if aPre == "" { // a is stable, b is pre
		return 1
	}
	if bPre == "" { // b is stable, a is pre
		return -1
	}
	// Simple prerelease lexical compare
	if aPre < bPre {
		return -1
	}
	if aPre > bPre {
		return 1
	}
	return 0
}

func splitPrerelease(v string) (main string, pre string) {
	parts := strings.SplitN(v, "-", 2)
	main = parts[0]
	if len(parts) > 1 {
		pre = parts[1]
	}
	return
}

func splitInts(v string) [3]int {
	var out [3]int
	segs := strings.Split(v, ".")
	for i := 0; i < 3 && i < len(segs); i++ {
		// ignore parse errors (treat as 0)
		n := 0
		for _, ch := range segs[i] {
			if ch < '0' || ch > '9' {
				n = 0
				break
			}
			n = n*10 + int(ch-'0')
		}
		out[i] = n
	}
	return out
}
