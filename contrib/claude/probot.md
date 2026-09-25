# Probot Architecture

## Layers

- `pkg/bot` is the domain-facing, provider-neutral contract. Its concrete
  service enqueues durable outbound messages in a caller-owned transaction, and
  it owns message intent contracts without importing Probot or a provider.
- `pkg/probot` implements bot dispatch, trusted run context, capability
  dispatch, and agent construction.
- `pkg/probot/capability/<domain>` owns provider-neutral message-type
  registration, actions, and agent tools. Neutral domain presentation and
  conversation parameters remain in `pkg/<domain>`.
- `pkg/probot/channel/<provider>` owns provider API clients, transport
  rendering, ingress, delivery queues, and reliability workers.
- Provider-feature bridge packages are prohibited. There must be no package,
  service, or composition object that combines a provider (such as Slack) with
  a feature (such as compliance portals). Capabilities emit provider-neutral
  messages and delivery targets. Provider adapters consume only those generic
  contracts. Composition roots may wire registries and interfaces, but must
  never branch on a provider-feature pair.

Dependencies flow from composition roots to adapters to narrow domain/core
contracts. Constructors receive required collaborators; do not add `Use*`
methods for required dependencies or wire registries after construction.

## Capability authoring

Every capability implements `probot.Capability` by providing a stable name.
Implement only the optional contributor contracts the capability needs:

- `MessageCapability` for message types and intent rendering.
- `OutboundMessageCapability` for provider-neutral notification cards.
- `ActionCapability` for action prefixes and action handling.
- `ToolContributor` for LLM tools.

Register capabilities once at startup. The registry is the canonical inbound
dispatch surface. Keep persisted metadata as a generic envelope, then decode
and validate it into capability-owned types before use. Tool handlers must
derive organization, conversation, message, and actor identifiers from trusted
run context rather than model arguments. State-changing tools need a stable
idempotency key whenever the run context provides one.

## Provider policy

Providers own transport clients, rendering, ingress, delivery, and channel-only
tools. They do not register with Probot through a provider prompt interface.
Do not add registries or speculative provider feature flags. Provider-specific
cards, limits, threading, reactions, and assistant status remain in the
provider package.

`bot.MessageIntent` is the provider-neutral outbound contract. Capabilities
produce fallback text, cards, and actions; the active channel adapter
transforms and durably queues that intent. Channel-only verbs such as Slack
reactions remain provider tools. Probot is one bot across channels: Slack,
future web chat, and other providers share capabilities and conversation
history while keeping transport policy in their adapters.

## Identity binding

`pkg/probot/identitybinding` owns external-account binding persistence,
opaque challenge issuance, confirmation, lookup, and deletion. A provider maps
its signed actor coordinates to an `identitybinding.Subject`; capabilities
receive only the resulting trusted `RunContext.IdentityID`.

Provider adapters own bind-required delivery and presentation. Slack posts a
generic `/probot login` prompt in the conversation (no token, no DM). The
Slackbot slash command Request URL is `https://<baseUrl>/slack/v1/commands`;
`/probot login` replies with an ephemeral Block Kit button whose URL contains
only a random hashed challenge. Each challenge is single-use: preview and
confirm fail after the token is consumed. After the user confirms while
signed in, Slack replaces that ephemeral prompt with a linked confirmation.
Never set
`response_type` to `in_channel` on that reply. Teams adaptive cards and email delivery must not move into the
identity-binding package. External user identifiers must never be embedded in
bind URLs.

## Legacy Slack boundary

Legacy compatibility belongs entirely to the provider adapter and is expressed
through provider-generic message lookup, action-alias, and delivery contracts.
Modern message lookup is always organization-scoped. An unscoped lookup is
permitted only for legacy interactive traffic and emits a structured warning.

Remove the compatibility implementation only after all of the following hold
for the agreed retention window:

1. No legacy Slack connectors remain.
2. No legacy messages are pending.
3. No legacy interactive fallback traffic is observed.

## Reliability workers

`probod` starts exactly one provider-neutral `agent-execution-worker`. It
schedules conversational executions from `agent_executions`/`agent_inputs`,
using the agent-profile and execution-adapter registries. GraphQL exposes
executions for observability and approval only; it does not create one-shot
runs. Conversational executions appear after someone `@mentions` the bot in a
channel, or after a DM `message` event (including DM edits). It also starts a
provider-neutral `probot-message-worker` that claims `bot_messages`, renders
via `OutboundMessageCapability`, and hands the intent to the channel adapter.
These workers run even when Slackbot is disabled; provider-specific ingress
and delivery workers only start when their provider is enabled. When Slackbot
is disabled, compliance workflows do not enqueue bot messages, so a
deployment without Slack remains fully functional without accumulating jobs
that have no registered adapter.

Slack ingress persists events before acknowledgement. Event, notification,
interactive-command, and delivery-operation workers use bounded concurrency,
database claims, stale-claim recovery, and bounded retries. Initial sends use
stable client message IDs.

All Probot workers receive the process Prometheus registerer and tracer
provider. The worker kit exports claim/task outcome and duration metrics with
the worker name as a bounded label. Worker logs add only opaque correlation
identifiers (`event_id`, `command_id`, `agent_id`, `message_id`, and
`operation_id`), attempt count, and terminal state; never add Slack payloads,
user IDs, response URLs, message text, or other user-supplied values.

## Transaction boundaries and state machines

The compliance-domain transaction that creates or changes an access request
calls `bot.Service.EnqueueMessage` with the `compliance_access` capability,
message type, access ID attribute, stable subject, and `POST` or `UPDATE`
purpose. The call inserts a `bot_messages` outbox row in the same transaction
when Slackbot is enabled. It is a no-op when Slackbot is disabled, and never
calls a provider or writes `slackbot_messages`. Notifications are not agent
turns: they do not create `agent_executions` or `agent_inputs`.

Install-time channel verification (`QueueWelcome`) is the exception: it is
not a domain notification, so it skips `bot_messages` and goes straight to
the Slack delivery queue (`slackbot_messages`) via `DeliverVerification`.
Replay uses the same `client_msg_id` as other modern Slack posts.

The bot-message worker claims an outbox row, asks the capability to build an
`OutboundMessage`, and the Slack adapter queues or updates `slackbot_messages`.
Successful first delivery upserts a `bot_thread_subjects` row so a later
`@mention` in that thread can create a conversational execution with the
right capability tools.

The global scheduler claims an eligible execution (`IDLE` or `SUSPENDED`)
with `FOR UPDATE SKIP LOCKED` and a random owner token. The provider adapter
revalidates the current installation, decorates the selected agent profile with
capability/channel tools, and supplies trusted run context from the USER
input's identity and per-turn `source_coordinates`. Trust is computed at
`Prepare` time from the live installation, binding, and `bot_thread_subjects`
row; it is not persisted on the execution. Only `USER` inputs exist. Each
conversational claim processes one USER input so concurrent pings in a shared
thread do not share the last writer's authorization. Slack tools bind
reactions and replies to that input's channel, thread, and message
timestamps. Assistant final text is never auto-delivered: conversational
channel-visible effects are queued by channel tools with stable operation
keys.

Slack ingress sets the assistant thread status indicator when it enqueues a
turn, and a run hook clears it when the turn ends without queueing a reply
(reaction-only turns, failures, approval stops). Slack clears the indicator
itself once the app posts in the thread, so a replying turn leaves it alone.
Both sides derive the indicator thread from the same turn coordinates so a
turn always clears the indicator it set.

Slack event HTTP ingress verifies the signature, validates the envelope, and
inserts `slackbot_events` before returning 200. A duplicate Slack event ID is
an acknowledged no-op. Channel agents start only on `@mention`. DMs start on
`message` events and re-trigger on DM edits. The Slack adapter fetches
`conversations.replies` outside the database transaction, formats a thread
snapshot, creates the execution if none exists for that thread, and enqueues
a USER input that stores its own `source_coordinates`. Channel message edits
and reactions do not start agents. DMs keep a per-user session. The event
worker transitions:

`pending -> processing -> processed`

or, on failure:

`processing -> pending(next_attempt_at) -> dead_lettered`.

`slackbot_processed_events` is the inner event-effect dedupe ledger. Keep both
layers: the inbox makes acknowledgement durable, while the ledger prevents a
replayed event from repeating committed event side effects.

Interactive HTTP ingress looks up the clicker's identity binding before
enqueue. Unbound actors are not queued. Slack block actions ignore most HTTP
response bodies, so the adapter posts an ephemeral `/probot login` prompt to
the click `response_url` (`replace_original: false`) and acknowledges Slack.
Bound traffic stores only an encrypted payload and request digest in
`slackbot_interactive_commands`, then acknowledges Slack. The worker decrypts
after claim, reloads the current installation, binding, and organization-scoped
message, and dispatches through the capability registry. Forbidden or invalid
actions also post an ephemeral error to the same `response_url` before
dead-lettering. Its states are the same pending/processing/processed or
retry/dead-letter states as the event inbox. Invalid ciphertext, malformed or
incomplete payloads, missing installations, missing or revoked bindings,
cross-workspace message lookups, forbidden actions, and unknown capabilities
are terminal failures.

Agent execution claims use a random owner token and a lease heartbeat. Every
checkpoint, input transition, and terminal write is fenced by that token. On
process shutdown an active run suspends to its durable checkpoint. A restarted
worker restores the checkpoint and reuses provider-issued tool-call IDs in
operation keys, so committed tools and queued deliveries are not repeated.

Provider notification messages and Slack delivery operations use the same
short claim, external call, and outcome-persist pattern. `slackbot_messages` is
solely the fully rendered modern Slack delivery queue; every queued row already
has a channel ID. The production `slack_messages` table and `pkg/slack` routing
remain the compatibility path for legacy Slack connectors and interactive
traffic until the retirement conditions below are met. Successful modern
notification delivery and its success hook commit together. A post-delivery
persistence failure clears the local sent marker and schedules replay with the
same client message ID. Reaction replay treats Slack's `already_reacted`
response as success.

## Operation-key contract

Every state-changing capability action or agent tool must provide a stable,
non-empty operation key. The key must identify the logical operation, not an
attempt:

- Slack actions use the digest of the signed interactive request.
- Agent tools derive the key from the organization, trusted conversation and
  message anchors, action, and selected resource.
- Provider delivery operations persist that key under the organization and
  use a unique `(organization_id, operation_key)` constraint.

The capability's business mutation and `operation_receipts.Claim` occur in the
same database transaction. If claim returns false, the prior mutation already
committed and the whole transactional effect is skipped. Queueing a provider
operation uses an organization-scoped upsert and returns the existing row on
replay. Never derive operation keys from retry counts, timestamps, model
output, response tokens, or mutable display data.

## Retry, dead letter, replay, and retention

Claims increment attempt count. Transient failures schedule bounded
exponential backoff; Slack rate-limit delays are honored up to the worker
maximum. Permanent validation, trust, and routing failures dead-letter
immediately. Exhausted transient failures also dead-letter. Dead letters are
retained for operator diagnosis and are not automatically replayed; replay
requires correcting the cause and explicitly resetting the row's processing
state and attempt budget while preserving its deduplication or operation key.

### Dead-letter replay (operators)

Fix the cause first (secret, destination, installation, payload). Then reset
**one row** in a transaction. Never change `operation_key`, `client_msg_id`,
`event_id`, `request_digest`, or `source_event_id`. Do not delete
`operation_receipts` or `slackbot_processed_events`; those ledgers make a
replay that already committed a no-op.

Inbox tables (`slackbot_events`, `slackbot_interactive_commands`,
`bot_messages`) share the same columns:

```sql
UPDATE slackbot_events
SET
    processing_started_at = NULL,
    processed_at = NULL,
    dead_lettered_at = NULL,
    next_attempt_at = NULL,
    attempt_count = 0,
    last_error = NULL,
    updated_at = NOW()
WHERE id = $1
  AND dead_lettered_at IS NOT NULL;
```

Use the same SET list for `slackbot_interactive_commands` and
`bot_messages` (match on `id`). Delivery operations use `completed_at`
instead of `processed_at`:

```sql
UPDATE slack_delivery_operations
SET
    processing_started_at = NULL,
    completed_at = NULL,
    dead_lettered_at = NULL,
    next_attempt_at = NULL,
    attempt_count = 0,
    last_error = NULL,
    updated_at = NOW()
WHERE id = $1
  AND dead_lettered_at IS NOT NULL;
```

Modern Slack notification rows terminal-fail on `error`, not
`dead_lettered_at`. Leave `client_msg_id` unchanged:

```sql
UPDATE slackbot_messages
SET
    processing_started_at = NULL,
    error = NULL,
    last_error = NULL,
    next_attempt_at = NULL,
    attempt_count = 0,
    updated_at = NOW()
WHERE id = $1
  AND sent_at IS NULL
  AND error IS NOT NULL;
```

Conversational agent executions and their USER inputs. Reset the execution
to `IDLE` (not `PENDING` or `COMPLETED`):

```sql
UPDATE agent_executions
SET
    status = 'IDLE',
    started_at = NULL,
    processing_owner_token = NULL,
    processing_heartbeat_at = NULL,
    processing_input_ids = '{}',
    next_attempt_at = NULL,
    last_error = NULL,
    error_message = NULL,
    dead_lettered_at = NULL,
    attempt_count = 0,
    updated_at = NOW()
WHERE id = $1
  AND dead_lettered_at IS NOT NULL;

UPDATE agent_inputs
SET
    processed_at = NULL,
    dead_lettered_at = NULL,
    next_attempt_at = NULL,
    attempt_count = 0,
    last_error = NULL,
    updated_at = NOW()
WHERE id = $2
  AND dead_lettered_at IS NOT NULL;
```

Dry-run the chosen UPDATE on staging and confirm the worker claims the row
before using it in production.

Completed reliability records and processed agent inputs are retained for 30
days. Dead letters and idle conversational executions are retained for 90
days; deleting an execution cascades its provider anchors and any remaining
inputs. Operation receipts are retained for 120 days, so they outlive every
checkpoint, command, outbox, dead-letter, and delivery replay horizon. The
hourly `probot-reliability-retention-worker` deletes at most 1,000 rows per
table per transaction.

## Legacy Slack retirement

Every compliance notification route emits an identifier-only structured log
with `backend=slackbot` or `backend=legacy-slack`; duplicate source-event
lookups emit the same backend field. Unscoped legacy interactive lookup emits
a dedicated warning. Operators can compare backend log counts and alert on
legacy fallback without exposing message contents or requester data.

Remove the legacy compatibility implementation only after one full retention
window in which all of these are true:

1. The database contains no legacy Slack connectors or pending legacy
   messages.
2. Routed notification logs contain no `backend=legacy-slack`.
3. No unscoped legacy interactive fallback warning is observed.
4. Modern queue depth, retry, dead-letter, and latency metrics are healthy.
5. A rollback window and explicit dead-letter replay procedure have been
   exercised.

## Production launch

Enable Slackbot in production only after the following are true:

1. Slackbot `signingSecret` is the Slackbot app secret. Do not copy the
   Slack connector signing secret. `/slack/v1/interactive` accepts both
   secrets during the dual-stack window; events and commands use Slackbot
   only.
2. Slack app URLs are `https://<baseUrl>/slack/v1/events`,
   `/slack/v1/interactive`, and `/slack/v1/commands`. Create a `/probot`
   slash command pointing at the commands URL (`/probot login`). OAuth must
   request the `commands` bot scope or `/probot` will not appear after
   install; reinstall the workspace after adding that scope.
3. Helm has `clientId`, `clientSecret`, `redirectUri`, and the LLM
   provider/model. Empty Slackbot secrets fail the chart when enabled.
   If you set a Slack connector `endpointToken`, it must come from the
   chart Secret (`secretKeyRef`), not a plain env value.
4. Coredata migrations through `20260814T115700Z` have been applied.
   Destructive drops (`input_messages`, `result`, `trusted_context`) and
   the `agent_runs` → `agent_executions` rename require the new binary to
   be fully rolled out before those migrations run. Status values
   `PENDING` and `COMPLETED` are rewritten to `IDLE`.
5. Alerts exist for dead letters, queue age, `backend=legacy-slack` log
   volume, and unscoped legacy interactive lookup warnings.
6. The dead-letter UPDATE above has been dry-run on staging and the worker
   reclaimed the row.
7. Legacy Slack stays until the retirement conditions in the previous
   section hold.
8. Slack ingress and the agent-execution worker cut over together so every
   pending `agent_inputs` row has `source_coordinates`. Inputs without
   coordinates fall back to the execution's latest message timestamps.
