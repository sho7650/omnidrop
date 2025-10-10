package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchScope(t *testing.T) {
	tests := []struct {
		name         string
		clientScope  string
		required     string
		expected     bool
		description  string
	}{
		{
			name:         "exact match",
			clientScope:  "tasks:write",
			required:     "tasks:write",
			expected:     true,
			description:  "Exact scope match should return true",
		},
		{
			name:         "wildcard match",
			clientScope:  "automation:*",
			required:     "automation:video-convert",
			expected:     true,
			description:  "Wildcard scope should match specific scope",
		},
		{
			name:         "wildcard match multiple levels",
			clientScope:  "automation:*",
			required:     "automation:notify",
			expected:     true,
			description:  "Wildcard should match any scope with same prefix",
		},
		{
			name:         "admin wildcard",
			clientScope:  "*",
			required:     "tasks:write",
			expected:     true,
			description:  "Admin wildcard should match everything",
		},
		{
			name:         "no match - different resource",
			clientScope:  "tasks:write",
			required:     "files:write",
			expected:     false,
			description:  "Different scopes should not match",
		},
		{
			name:         "no match - different action",
			clientScope:  "tasks:read",
			required:     "tasks:write",
			expected:     false,
			description:  "Different actions should not match",
		},
		{
			name:         "no match - wildcard wrong prefix",
			clientScope:  "tasks:*",
			required:     "files:write",
			expected:     false,
			description:  "Wildcard should not match different resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchScope(tt.clientScope, tt.required)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestHasRequiredScopes(t *testing.T) {
	tests := []struct {
		name           string
		clientScopes   []string
		requiredScopes []string
		expected       bool
		description    string
	}{
		{
			name:           "single scope exact match",
			clientScopes:   []string{"tasks:write"},
			requiredScopes: []string{"tasks:write"},
			expected:       true,
			description:    "Client with exact required scope should pass",
		},
		{
			name:           "multiple scopes all match",
			clientScopes:   []string{"tasks:write", "files:write"},
			requiredScopes: []string{"tasks:write", "files:write"},
			expected:       true,
			description:    "Client with all required scopes should pass",
		},
		{
			name:           "wildcard covers required scope",
			clientScopes:   []string{"automation:*"},
			requiredScopes: []string{"automation:video-convert"},
			expected:       true,
			description:    "Wildcard scope should cover specific scope",
		},
		{
			name:           "multiple wildcards",
			clientScopes:   []string{"tasks:*", "automation:*"},
			requiredScopes: []string{"tasks:write", "automation:notify"},
			expected:       true,
			description:    "Multiple wildcards should cover multiple scopes",
		},
		{
			name:           "admin wildcard",
			clientScopes:   []string{"*"},
			requiredScopes: []string{"tasks:write", "files:write", "automation:notify"},
			expected:       true,
			description:    "Admin wildcard should cover all scopes",
		},
		{
			name:           "missing one required scope",
			clientScopes:   []string{"tasks:write"},
			requiredScopes: []string{"tasks:write", "files:write"},
			expected:       false,
			description:    "Client missing one required scope should fail",
		},
		{
			name:           "completely different scopes",
			clientScopes:   []string{"tasks:read"},
			requiredScopes: []string{"files:write"},
			expected:       false,
			description:    "Client with unrelated scopes should fail",
		},
		{
			name:           "empty client scopes",
			clientScopes:   []string{},
			requiredScopes: []string{"tasks:write"},
			expected:       false,
			description:    "Client with no scopes should fail",
		},
		{
			name:           "empty required scopes",
			clientScopes:   []string{"tasks:write"},
			requiredScopes: []string{},
			expected:       true,
			description:    "No required scopes should always pass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasRequiredScopes(tt.clientScopes, tt.requiredScopes)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}
