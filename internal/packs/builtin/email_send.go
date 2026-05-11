// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 The helmdeck contributors

package builtin

import (
	"github.com/tosin2013/helmdeck/internal/packs"
	"github.com/tosin2013/helmdeck/internal/security"
	"github.com/tosin2013/helmdeck/internal/vault"
)


func EmailSend(c *vault.Store, eg *security.EgressGuard) *packs.Pack {
	return &packs.Pack{
		Name: "email.send",
		Version: "v1", // where does version come from?
		Description: "Send a transactional email (to, subject, html/text, attachment)",
		InputSchema: packs.BasicSchema{ 
			Required: []string{"to"},
			Properties: map[string]string{
				"to": "string",
			},
		},
		OutputSchema: packs.BasicSchema{
			Required: []string{"message_id"},
			Properties: map[string]string{
				"message_id": "string",
			},
		},
		// Handler: emailSendHandler(v, eg),
	}
}

type emailSendInput struct {
	To string `json:"to"`
}