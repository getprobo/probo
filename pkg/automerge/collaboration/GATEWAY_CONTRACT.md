# Repo collaboration gateway contract

This is the integration contract for migrating Probo's real-time document editing
from the custom `automerge-sync-v1` WebSocket protocol to the official
`@automerge/automerge-repo` protocol. It pins the three decisions the gateway and
the frontend must agree on — **document id**, **auth**, and **seeding** — plus the
rollout shape. The transport-agnostic protocol drivers it builds on
(`ServerConn`, `ClientConn`), the hub fan-out primitives, and the presence value
already exist and are tested; see [`PROTOCOL.md`](PROTOCOL.md).

The custom client being replaced lives in
`apps/console/src/pages/organizations/documents/description/_lib/AutomergeDocumentHandle.ts`
(transport) and `packages/ui/src/RichEditor/*` (the ProseMirror ↔ Automerge
binding and presence decorations). The binding layer is reusable as long as the
new client still exposes an `@automerge/prosemirror` `DocHandle`; only the
transport is replaced.

## 1. Routing and scope

Keep the connection **version-scoped by URL path**, exactly as the custom route
is:

```
{ws|wss}://{host}/api/console/v1/document-versions/{documentVersionID}/repo
```

- The `documentVersionID` path segment (a Probo `DocumentVersion` GID) is the
  single source of truth for **authorization** and **room selection**. It is not
  the repo document id.
- A new `/repo` route is added next to the existing `/sync` route so the two
  protocols coexist during migration. Mount it in the same authenticated router
  group in `pkg/server/api/console/v1/resolver.go`.
- One WebSocket serves exactly one document version, matching the hub's
  one-room-per-version model (`hub.acquire(scope, documentVersionID, ...)`).

Rationale: the repo protocol has no notion of "which Probo resource is this"; the
path keeps auth and room routing on the proven code path and out of the protocol.

## 2. Document id

The repo document id is **derived deterministically from the version GID**, not
chosen freely:

```
documentID  = base58check( sha256(documentVersionID)[:16] )   // DeriveDocumentID
automergeURL = "automerge:" + documentID                       // DeriveAutomergeURL
```

- `DeriveDocumentID` / `DeriveAutomergeURL` (this package, `documentid.go`) are
  the canonical Go implementation; the frontend must implement the identical
  derivation in TypeScript (same SHA-256, first 16 bytes, bs58check). The Go
  codec is validated against a real `@automerge/automerge-repo` id, so the two
  will agree.
- **Determinism is required, not cosmetic.** All peers of a version must use the
  same id because ephemeral gossip (presence, cursors) is keyed by document id: a
  peer silently drops an ephemeral frame whose id it does not recognise. Sync
  alone would tolerate differing ids (the server re-tags per connection), but
  presence would not.
- The server needs **no** id-derivation logic: it uses `NewAdoptingServerConn`,
  which binds to whatever id the client requests in its first sync/request frame
  and answers for that id (rejecting any second id on the connection). Go agents
  that want to open a version compute the URL with `DeriveAutomergeURL(gid)`.

Rationale: derivation gives every browser tab and Go agent the same id with zero
coordination and no new storage, and it makes presence line up.

## 3. Authentication

Auth is **unchanged** from the custom route; the repo `PeerId` is never trusted.

- **Browser** — a same-origin WebSocket upgrade carries the session cookie
  automatically. The `/repo` route sits behind the existing middleware
  (`NewSessionMiddleware`, `NewAPIKeyMiddleware`, `NewOAuth2AccessTokenMiddleware`,
  `NewIdentityPresenceMiddleware`) and the same `authorize` check
  (`DocumentVersionGet` + `DocumentUpdate`). No token is placed in the URL.
- **Go agents / non-browser clients** — present an API key or bearer token as an
  `Authorization` header on the upgrade request (the WebSocket dialer sets request
  headers). The same middleware validates it.
- The authenticated identity is bound to the connection server-side and used for
  audit and presence attribution. The repo `senderId` in the `join` frame is
  peer-chosen and is treated purely as a routing label, never as identity.

Rationale: reuses the working auth path and keeps credentials out of URLs and out
of the peer-chosen protocol fields.

## 4. Seeding

The server is the **document authority and owns seeding**; there is no seed
field in the repo handshake, and clients never seed via the protocol.

- A collaboration connection is served a materialized Automerge document for the
  version. The custom `needsSeed` / `seedContent` / seed-owner handshake and
  `ReleaseCollaborationSeed` machinery are **dropped** from the repo path.
- **Implementation (done)** — the conversion was ported to Go
  (`pkg/automerge/prosemirror.ToSpans`, the inverse of `Render`, validated to
  round-trip the entire shared ProseMirror corpus). The `/repo` handler seeds
  lazily on first open: the connection that claims the seed converts the
  version's stored ProseMirror JSON to spans, writes them into the shared
  document, and persists; the persist marks the state seeded, so later
  connections skip it. No JavaScript build step and no draft-creation change are
  required.

Rationale: server-authoritative seeding matches the CRDT authority model and
removes fragile handshake state; the only real work is where the one-time
PM → spans conversion runs.

## 5. Presence and cursors

- Presence rides repo **ephemeral gossip**, not a side JSON channel. The server
  returns a non-duplicate ephemeral frame from `ServerConn.Receive` as `fanout`
  and publishes it with `lease.BroadcastEphemeral`; it writes other peers' frames
  from `lease.Ephemeral` to the socket.
- A caret/selection is a presence `update` whose value is a `TextSelectionValue`
  (this package): the addressed text field plus **stable Automerge cursors** for
  anchor and head, replacing the custom integer `anchorPosition`/`headPosition`.
  Cursors survive concurrent edits; offsets do not. The frontend builds them with
  `Text.getCursor` and resolves them with `Text.getCursorPosition`; the presence
  decorations in `packages/ui/src/RichEditor/presence.ts` render from the resolved
  positions.
- **Cross-instance ephemeral** is delivered by publishing the frame over the
  collaboration `NOTIFY` channel in a typed envelope
  (`realtime.CollaborationEphemeral`) alongside the bare version-id "changed"
  signal. The `/repo` loop relays each gossiped frame with
  `DocumentService.NotifyCollaborationEphemeral`; the receiving instance decodes
  the envelope in `notifyExternal` and fans it out to local peers, skipping the
  publishing instance's own echo. Oversized frames fall back to local-only
  fan-out.

## 6. Reconnect and revision

- The repo network adapter owns reconnect/backoff and the sync generation model,
  so the custom `revision` / `ready` / initialization-timeout handshake fields are
  dropped from the client.
- The server keeps its debounced persistence and its cross-instance sync refresh
  (`lease.SchedulePersist`, `lease.Wake`, `RefreshCollaboration`) unchanged; those
  are independent of the wire protocol.

## 7. Rollout

1. Add the `/repo` route wired to `NewAdoptingServerConn` + the hub, coexisting
   with `/sync`.
2. Add the TS `deriveDocumentId` helper (mirroring `DeriveDocumentID`) and a repo
   `NetworkAdapter` targeting `/repo`, behind a feature flag, reusing the existing
   `DocHandle`-based binding.
3. Validate with a Postgres-backed integration test and the live JS interop
   harness, then flip the flag.
4. Remove `/sync`, `AutomergeDocumentHandle.ts`, and the DB presence tables once
   no client uses the custom protocol.

## Status

| Piece | State |
|---|---|
| `ServerConn` / `ClientConn` drivers | done, tested |
| Adopting document id (`NewAdoptingServerConn`) | done, tested |
| Document id derivation + `automerge:` URL (`documentid.go`) | done, tested against a real repo id |
| Opaque ephemeral fan-out (`BroadcastEphemeral` / `Ephemeral`) | done, tested |
| Cursor-based selection presence (`TextSelectionValue`) | done, tested |
| TS `deriveDocumentId` mirror (`@probo/ui`) | done, tested for byte-parity with Go |
| TS cursor-based selection helper (`@probo/ui` `repoSelection`) | done, tested for cursor stability across concurrent edits (browser↔browser) |
| `/repo` route wiring | done, driven end-to-end in Go by a real `ClientConn` (Postgres integration + live JS still pending) |
| ProseMirror → spans forward conversion (`prosemirror.ToSpans`) | done, round-trips the whole shared corpus |
| Server-authoritative seeding | done, seeds on first open from stored content and materializes over the loop (Postgres lifecycle still to integration-test) |
| Cross-instance ephemeral fan-out | done, published over the NOTIFY channel and delivered to local peers with self-echo suppression (Postgres delivery still to integration-test) |
| Frontend repo client (`connectRepoDocument`) | done, `Repo` + `WebSocketClientAdapter` over `/repo`; presence rides repo ephemeral with stable cursors (PM↔Automerge mapping unit-tested) |
| Legacy `/sync` removal | done: `/sync` route, the custom `AutomergeDocumentHandle`, the DB-backed presence, and the hub's structured presence are removed; the editor uses `/repo` exclusively |
