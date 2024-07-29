package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SafeBrowsingRequest struct {
	Client struct {
		ClientID      string `json:"clientId"`
		ClientVersion string `json:"clientVersion"`
	} `json:"client"`
	ThreatInfo struct {
		ThreatTypes      []string `json:"threatTypes"`
		PlatformTypes    []string `json:"platformTypes"`
		ThreatEntryTypes []string `json:"threatEntryTypes"`
		ThreatEntries    []struct {
			URL string `json:"url"`
		} `json:"threatEntries"`
	} `json:"threatInfo"`
}

type SafeBrowsingResponse struct {
	Matches []struct {
		ThreatType      string `json:"threatType"`
		PlatformType    string `json:"platformType"`
		ThreatEntryType string `json:"threatEntryType"`
		Threat          struct {
			URL string `json:"url"`
		} `json:"threat"`
	} `json:"matches"`
}

// todo: retries should be shifted to ActionRunner
const maxRetries = 3
const initialRetryDelay = 2 * time.Second

func SafeBrowsingCheckURL(apiKey, apiUrl, urlToCheck string) (*SafeBrowsingResponse, error) {
	apiURL := apiUrl + "?key=" + apiKey

	requestPayload := SafeBrowsingRequest{}
	requestPayload.Client.ClientID = "stealth_project"
	requestPayload.Client.ClientVersion = "1.0"
	requestPayload.ThreatInfo.ThreatTypes = []string{"MALWARE", "SOCIAL_ENGINEERING"}
	requestPayload.ThreatInfo.PlatformTypes = []string{"ANY_PLATFORM"}
	requestPayload.ThreatInfo.ThreatEntryTypes = []string{"URL"}
	requestPayload.ThreatInfo.ThreatEntries = append(requestPayload.ThreatInfo.ThreatEntries, struct {
		URL string `json:"url"`
	}{URL: urlToCheck})

	jsonPayload, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, err
	}

	var response SafeBrowsingResponse
	var retryDelay = initialRetryDelay

	for attempts := 0; attempts < maxRetries; attempts++ {
		resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			if attempts < maxRetries-1 {
				time.Sleep(retryDelay)
				retryDelay *= 2
				continue
			} else {
				return nil, fmt.Errorf("rate limit exceeded, max retries reached")
			}
		}

		// Check for other non-success status codes.
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		// Parse the response JSON.
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, err
		}

		return &response, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts", maxRetries)
}
