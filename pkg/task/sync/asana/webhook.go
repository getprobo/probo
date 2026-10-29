// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package asana

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const (
	HeaderHookSecret    = "X-Hook-Secret"
	HeaderHookSignature = "X-Hook-Signature"
)

type (
	Event struct {
		Action   string `json:"action"`
		Resource struct {
			GID             string `json:"gid"`
			ResourceType    string `json:"resource_type"`
			ResourceSubtype string `json:"resource_subtype"`
		} `json:"resource"`
		Parent struct {
			GID          string `json:"gid"`
			ResourceType string `json:"resource_type"`
		} `json:"parent"`
	}

	EventBatch struct {
		Events []Event `json:"events"`
	}
)

func Sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)

	return hex.EncodeToString(mac.Sum(nil))
}

func VerifySignature(secret, signature string, body []byte) bool {
	if secret == "" || signature == "" {
		return false
	}

	expected := Sign(secret, body)

	return hmac.Equal([]byte(expected), []byte(signature))
}

func ParseEvents(body []byte) (*EventBatch, error) {
	var batch EventBatch
	if err := json.Unmarshal(body, &batch); err != nil {
		return nil, fmt.Errorf("cannot unmarshal Asana webhook: %w", err)
	}

	return &batch, nil
}

func (e Event) TaskGID() string {
	if e.Resource.ResourceType == "task" {
		return e.Resource.GID
	}

	if e.Parent.ResourceType == "task" {
		return e.Parent.GID
	}

	return ""
}

func (e Event) DeletedTask() bool {
	return e.Resource.ResourceType == "task" && (e.Action == "deleted" || e.Action == "removed")
}

func (e Event) AddedTask(projectGID string) bool {
	return e.Action == "added" &&
		e.Resource.ResourceType == "task" &&
		e.Parent.ResourceType == "project" &&
		e.Parent.GID == projectGID &&
		e.Resource.GID != ""
}
