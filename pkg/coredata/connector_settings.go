// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package coredata

import (
	"encoding/json"
	"fmt"
)

type (
	SlackConnectorSettings struct {
		Channel   string `json:"channel,omitempty"`
		ChannelID string `json:"channel_id,omitempty"`
	}

	TallyConnectorSettings struct {
		OrganizationID string `json:"organization_id"`
	}

	OnePasswordConnectorSettings struct {
		SCIMBridgeURL string `json:"scim_bridge_url"`
	}
)

// SetSettings marshals a typed settings struct into the connector's RawSettings.
func (c *Connector) SetSettings(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("cannot marshal connector settings: %w", err)
	}
	c.RawSettings = data
	return nil
}

// SlackSettings unmarshals the connector's RawSettings into SlackConnectorSettings.
func (c *Connector) SlackSettings() (SlackConnectorSettings, error) {
	var s SlackConnectorSettings
	if err := c.unmarshalSettings(&s); err != nil {
		return s, err
	}
	return s, nil
}

// TallySettings unmarshals the connector's RawSettings into TallyConnectorSettings.
func (c *Connector) TallySettings() (TallyConnectorSettings, error) {
	var s TallyConnectorSettings
	if err := c.unmarshalSettings(&s); err != nil {
		return s, err
	}
	return s, nil
}

// OnePasswordSettings unmarshals the connector's RawSettings into OnePasswordConnectorSettings.
func (c *Connector) OnePasswordSettings() (OnePasswordConnectorSettings, error) {
	var s OnePasswordConnectorSettings
	if err := c.unmarshalSettings(&s); err != nil {
		return s, err
	}
	return s, nil
}

func (c *Connector) unmarshalSettings(v any) error {
	if len(c.RawSettings) == 0 || string(c.RawSettings) == "null" {
		return nil
	}
	if err := json.Unmarshal(c.RawSettings, v); err != nil {
		return fmt.Errorf("cannot unmarshal connector settings: %w", err)
	}
	return nil
}
