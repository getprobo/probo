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

package bot

import (
	"context"
	"time"

	"go.probo.inc/probo/pkg/gid"
)

type (
	ActionStyle string

	// ActionIntent is a single control. When Options is set the control offers
	// those choices instead of firing on click, and the chosen option value is
	// reported back as the selected value.
	ActionIntent struct {
		ID      string
		Label   string
		Style   ActionStyle
		Value   string
		URL     string
		Options []ActionOptionIntent
	}

	ActionOptionIntent struct {
		Label string
		Value string
	}

	// ItemIntent is one row of a group: a label that may link to the resource,
	// an optional state, and at most one control.
	ItemIntent struct {
		ID     string
		Label  string
		URL    string
		Status string
		Action *ActionIntent
	}

	GroupIntent struct {
		ID    string
		Title string
		Items []ItemIntent
	}

	MessageIntent struct {
		FallbackText string
		Headline     string
		Context      string
		Body         string
		Actions      []ActionIntent
		Groups       []GroupIntent
	}

	Message struct {
		ID             gid.GID
		OrganizationID gid.GID
		Type           string
		Attributes     map[string]any
	}

	DeliveredMessage struct {
		Message   Message
		CreatedAt time.Time
		Backend   string
	}

	MessageRenderer interface {
		RenderMessage(ctx context.Context, message Message) (MessageIntent, error)
	}
)

const (
	ActionStylePrimary ActionStyle = "primary"
	ActionStyleDanger  ActionStyle = "danger"
)
