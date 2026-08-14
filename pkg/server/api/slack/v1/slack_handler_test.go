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

package slack_v1

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
)

func TestSlackInteractiveAcknowledgesAfterDurableInsert(t *testing.T) {
	t.Parallel()

	inbox := newInteractiveInbox(t)
	rawPayload := uniqueInteractivePayload(t)
	response := performSlackAction(t, inbox, rawPayload)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{"success":true}`, response.Body.String())

	command := loadInteractiveCommand(t, inbox, rawPayload)
	assert.Nil(t, command.ProcessedAt)
	assert.Nil(t, command.DeadLetteredAt)
}

func TestSlackInteractiveAcknowledgesDuplicate(t *testing.T) {
	t.Parallel()

	inbox := newInteractiveInbox(t)
	rawPayload := uniqueInteractivePayload(t)
	first := performSlackAction(t, inbox, rawPayload)
	second := performSlackAction(t, inbox, rawPayload)

	assert.Equal(t, http.StatusOK, first.Code)
	assert.Equal(t, http.StatusOK, second.Code)
	assert.JSONEq(t, `{"success":true}`, second.Body.String())
}

func TestSlackInteractiveRejectsMalformedPayload(t *testing.T) {
	t.Parallel()

	response := performSlackAction(t, newInteractiveInbox(t), `{"user":{"id":"U123"}}`)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Body.String(), "cannot parse Slack payload")
}

func performSlackAction(
	t *testing.T,
	inbox *slackchannel.InteractiveCommandInbox,
	rawPayload string,
) *httptest.ResponseRecorder {
	t.Helper()

	digest := sha256.Sum256([]byte(rawPayload))
	t.Cleanup(func() { deleteInteractiveCommand(t, test.PGClient(t), digest[:]) })

	body := url.Values{"payload": []string{rawPayload}}.Encode()
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/slack/v1/interactive",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response := httptest.NewRecorder()
	SlackHandler(inbox, log.NewLogger()).ServeHTTP(response, req)

	return response
}

func newInteractiveInbox(t *testing.T) *slackchannel.InteractiveCommandInbox {
	t.Helper()

	client := test.PGClient(t)
	if !interactiveCommandsAvailable(t, client) {
		t.Skip("slackbot_interactive_commands is unavailable in the test database")
	}

	return slackchannel.NewInteractiveCommandInbox(
		client,
		cipher.EncryptionKey{1, 2, 3},
	)
}

func uniqueInteractivePayload(t *testing.T) string {
	t.Helper()

	id := gid.New(gid.NilTenant, coredata.SlackbotInteractiveCommandEntityType)

	return fmt.Sprintf(
		`{"team":{"id":"T-%s"},"user":{"id":"U123"},"container":{"channel_id":"C123","message_ts":"123.456"},"actions":[{"action_id":"accept_all","action_ts":"123.789","value":"message-id"}]}`,
		id,
	)
}

func loadInteractiveCommand(
	t *testing.T,
	inbox *slackchannel.InteractiveCommandInbox,
	rawPayload string,
) coredata.SlackbotInteractiveCommand {
	t.Helper()

	digest := sha256.Sum256([]byte(rawPayload))
	client := test.PGClient(t)

	var command coredata.SlackbotInteractiveCommand
	require.NoError(
		t,
		client.WithConn(
			context.Background(),
			func(ctx context.Context, conn pg.Querier) error {
				return command.LoadByRequestDigest(ctx, conn, digest[:])
			},
		),
	)

	return command
}

func deleteInteractiveCommand(t *testing.T, client *pg.Client, digest []byte) {
	t.Helper()

	_ = client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(
				ctx,
				`DELETE FROM slackbot_interactive_commands WHERE request_digest = $1`,
				digest,
			)

			return err
		},
	)
}

func interactiveCommandsAvailable(t *testing.T, client *pg.Client) bool {
	t.Helper()

	err := client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			var command coredata.SlackbotInteractiveCommand

			return command.LoadByRequestDigest(ctx, conn, []byte("missing"))
		},
	)

	return err == nil || errors.Is(err, coredata.ErrResourceNotFound)
}
