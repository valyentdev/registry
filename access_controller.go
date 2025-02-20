package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/distribution/distribution/v3/registry/auth"
	"golang.org/x/crypto/bcrypt"
)

type accessController struct {
	apiBaseURL    string
	ravelPassword []byte
}

func newAccessController(options map[string]any) (auth.AccessController, error) {
	password := os.Getenv("REGISTRY_RAVEL_PASSWORD")
	bcrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing the password: %w", err)
	}

	return &accessController{
		apiBaseURL:    os.Getenv("VALYENT_API_BASE_URL"),
		ravelPassword: bcrypted,
	}, nil
}

func (ac *accessController) Authorized(req *http.Request, accessRecords ...auth.Access) (*auth.Grant, error) {
	username, password, ok := req.BasicAuth()
	if !ok {
		return nil, &challenge{
			err: auth.ErrInvalidCredential,
		}
	}
	// When Ravel consumes the registry,
	// check if the password matches the expected one
	if username == "valyent" {
		// constant-time comparison
		if err := bcrypt.CompareHashAndPassword(ac.ravelPassword, []byte(password)); err != nil {
			return nil, &challenge{
				err: auth.ErrInvalidCredential,
			}
		} else {
			return &auth.Grant{
				User: auth.UserInfo{Name: "valyent"},
			}, nil
		}
	}

	// Check if the password matches a valid API key
	valid, err := ac.validateApiKey(password, accessRecords)
	if err != nil {
		return nil, &challenge{
			err: fmt.Errorf("error validating API key: %v", err),
		}
	}

	if !valid {
		return nil, &challenge{
			err: auth.ErrInvalidCredential,
		}
	}

	return &auth.Grant{
		User: auth.UserInfo{Name: "valyent"},
	}, nil
}

func (ac *accessController) validateApiKey(apiKey string, accessRecords []auth.Access) (bool, error) {
	// Create the request to the Valyent API
	url := fmt.Sprintf("%s/auth/api/check", ac.apiBaseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	// Set the Authorization header with the Bearer token
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Parse the response
	var result struct {
		Authenticated    bool   `json:"authenticated"`
		OrganizationSlug string `json:"organizationSlug"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	if !result.Authenticated {
		return false, nil
	}

	for _, accessRecord := range accessRecords {
		if !strings.HasPrefix(accessRecord.Name, result.OrganizationSlug) {
			return false, nil
		}
	}

	return true, nil
}
