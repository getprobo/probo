# automerge-repo protocol inventory

This package makes the Probo Go server and Go agents speak the same collaboration
protocol as the JavaScript `@automerge/automerge-repo` client, while keeping our
own pure-Go CRDT engine (`pkg/automerge`) as the document authority.

Everything here is derived from the pinned upstream source, not from guesswork.
The wire format upstream ships is **alpha**, so the pinned version is the
contract and the JS oracle fixtures under `testdata/` are the ground truth. No
Go decoder or encoder is trusted until it round-trips those fixtures.

## Pinned versions

| Package | Version | Role |
|---|---|---|
| `@automerge/automerge` | `3.4.1` (`^3.4.0`) | CRDT core; sync message and cursor bytes |
| `@automerge/automerge-repo` | `2.6.0-alpha.3` (exact) | repo message union, ephemeral gossip, Presence |
| `@automerge/automerge-repo-network-websocket` | to pin in the transport phase | WebSocket wire framing and join/peer/leave handshake |

The websocket adapter is a separate package and defines the actual on-socket
framing. It is intentionally out of scope for this first inventory, which covers
the message layer and Presence; the transport phase pins and inventories it.

## Layering

A byte on the socket nests three independently versioned layers:

```text
WebSocket adapter frame        (join / peer / leave + CBOR of the repo message)
  └── repo message             (sync | request | ephemeral | doc-unavailable | ...)
        └── payload
              ├── sync/request: an Automerge V2 sync message (our engine owns this)
              └── ephemeral:    CBOR of the Presence envelope
```

Our engine already owns the innermost layer. This package adds the middle layer
and Presence; the transport phase adds the outermost.

## Repo message union

From `dist/network/messages.d.ts`. Every message carries `senderId` and
`targetId` (both `PeerId`, a string). Type-specific fields:

| `type` | Fields | Notes |
|---|---|---|
| `sync` | `documentId`, `data: Uint8Array` | `data` is an Automerge sync message |
| `request` | `documentId`, `data: Uint8Array` | initial sync asking whether a peer has the doc |
| `ephemeral` | `documentId`, `sessionId`, `count: number`, `data: Uint8Array` | gossiped; not persisted |
| `doc-unavailable` | `documentId` | neither the peer nor its peers have the doc |
| `remote-subscription-change` | `add?`, `remove?` (`StorageId[]`) | storage-head subscription control |
| `remote-heads-changed` | `documentId`, `newHeads` | per-storage heads with timestamps |

For our use case the document-scoped subset (`DocMessage`) is what matters:
`sync`, `request`, `ephemeral`, `doc-unavailable`. The two remote-heads messages
support cross-storage head subscription, which our single-authority server does
not need initially; the gateway may ignore them (documented, not silently).

## Ephemeral gossip and de-duplication

Ephemeral messages are forwarded ("gossiped") to peers, so the protocol dedupes
by `(sessionId, count)`:

- `sessionId` is a random id chosen by a sender at startup.
- `count` is a per-sender sequence number that strictly increases.
- A receiver discards any `(sessionId, count)` it has already seen, which breaks
  forwarding loops.

The gateway must preserve `senderId`, `sessionId`, and `count` unchanged when
relaying, and must apply the same de-duplication before re-broadcasting.

## Presence envelope

Presence (`dist/presence/`) rides entirely inside the ephemeral `data`. It never
touches document history. The `data` bytes are the CBOR encoding of a single-key
envelope:

```jsonc
{ "__presence": <presence message> }
```

The marker key is `__presence` (`PRESENCE_MESSAGE_MARKER`). The four presence
messages:

| `type` | Fields | Meaning |
|---|---|---|
| `update` | `channel: string`, `value: any` | one channel's state changed |
| `snapshot` | `state: any` | full multi-channel state (sent on start and to newcomers) |
| `heartbeat` | — | liveness when nothing changed |
| `goodbye` | — | sender is leaving; forget it immediately |

Defaults (`dist/presence/constants.js`):

- heartbeat interval: `15000 ms`
- peer TTL: `45000 ms` (three missed heartbeats)

A peer is pruned once `peerTtlMs` passes with no message; `goodbye` prunes
immediately. `value`/`state` are application-defined; our documents will carry an
Automerge `Cursor` (opaque bytes) for selections rather than integer offsets.

## CBOR encoding

The ephemeral payload uses `cbor-x` configured as:

```js
new Encoder({ tagUint8Array: false, useRecords: false })
```

This matters for byte-level parity, and the Go codec must match it:

- `useRecords: false` — plain CBOR maps (major type 5), not cbor-x record
  extensions. No custom tags for object shapes.
- `tagUint8Array: false` — a `Uint8Array` is a plain CBOR byte string (major
  type 2), not wrapped in a typed-array tag.
- Map key order follows insertion order of the JS object; a tolerant decoder
  must not depend on key order, and the encoder should reproduce upstream order
  where a fixture asserts byte identity.

## Fixtures

`testdata/` holds JS-generated fixtures produced by
`packages/automerge-conformance/generate-collaboration-fixtures.mjs` using the
pinned packages' own CBOR encoder, so they are byte-exact:

- `presence-*.json` — each presence message type: the JS envelope plus its
  base64 CBOR `data` bytes.
- `ephemeral-*.json` — a full ephemeral repo message wrapping a presence payload.

Each Go codec change must round-trip these. The wire-framing fixtures (join/peer
handshake, socket frames) are added in the transport phase alongside the pinned
websocket adapter.

## Transport layer (WebSocket adapter)

Pinned: `@automerge/automerge-repo-network-websocket@2.6.0-alpha.3`. The current
`WebSocketClientAdapter` and server adapters encode every frame with the same
repo CBOR helper used for payloads (`useRecords: false`, `tagUint8Array: false`).
The older `encoder.js` and `WSShared.js` in that package are legacy compat and
are not used by the current adapters.

Each binary WebSocket frame is exactly one CBOR-encoded message; the WebSocket
message boundary is the framing, so there is no length prefix of our own. The
server must read binary frames (not text) and treat each as one message.

Protocol version is `"1"` (`ProtocolV1`).

### Handshake

```text
client ──▶ join   { type:"join", senderId, peerMetadata, supportedProtocolVersions:["1"] }
server ──▶ peer   { type:"peer", senderId, targetId, peerMetadata, selectedProtocolVersion:"1" }
```

- `join` is the first frame the client sends, before it knows the server peer id,
  so it has no `targetId`.
- The server replies `peer` selecting a protocol version, or `error`
  `{ type:"error", senderId, targetId, message }` and then closes the socket.
- After the handshake, both directions exchange the repo messages from the
  message-union section (`sync`, `request`, `ephemeral`, `doc-unavailable`, and
  the two remote-heads messages), each as its own CBOR frame.
- There is no explicit `leave` frame; a disconnect is the socket closing. A
  presence `goodbye` (inside an ephemeral) is the graceful application-level
  signal.

### PeerMetadata

```text
{ storageId?: string (StorageId), isEphemeral?: boolean }
```

Both fields are optional. `isEphemeral` marks a peer that does not persist
documents. Our gateway can present its own metadata and must not trust a peer's
metadata as identity (see below).

### Sync choreography (server as authority)

The server WebSocket adapter is a pure relay: it performs the handshake and then
hands every message to the repo's synchronizer. The synchronizer, not the
adapter, runs the sync protocol, so a gateway that is itself the document
authority must reproduce the synchronizer's server-side behavior. From
`DocSynchronizer`:

- Inbound `sync` and `request` are handled identically: apply the payload with
  `receiveSyncMessage`, then generate outbound messages.
- The message type the synchronizer emits is `request` only when it does **not**
  have the document (no heads, empty shared heads, peer status unknown);
  otherwise it emits `sync`. Our gateway always holds the document, so it
  **always emits `sync`** and never `request`.
- `doc-unavailable` is emitted only when the responder has no data and
  availability settles unavailable. Our gateway always has the document (access
  is decided by authentication at connect time, returning 404/403 rather than a
  protocol frame), so it does **not** send `doc-unavailable`.
- The payload bytes are exactly `generateSyncMessage`/`receiveSyncMessage` from
  `@automerge/automerge`, which is the same V2 sync protocol
  `pkg/automerge.SyncState` implements, so repo `sync.data` is one of our sync
  messages unchanged. This is the interop linchpin and is covered by the existing
  sync parity suite.

The resulting server loop: on connect, announce by draining
`GenerateMessage` into `sync` frames; on each inbound `sync`/`request`, call
`ReceiveMessage` then drain `GenerateMessage` into `sync` frames; forward
non-duplicate `ephemeral` frames to the room and room frames to the socket.

### Gateway responsibilities (transport)

- Accept a binary WebSocket, read `join`, negotiate `"1"`, reply `peer` with a
  server `senderId`.
- The repo `PeerId` in `join` is peer-chosen and is **not** a user identity;
  authenticate the connection out of band (our existing session auth) and bind
  the authenticated identity to the connection, never to `senderId`.
- Route `sync`/`request` payloads into the per-peer, per-document
  `automerge.SyncState`; forward `ephemeral` frames to the room with the existing
  cross-instance fanout, de-duplicated by `(sessionId, count)`.

## Fixtures

`testdata/` also holds transport fixtures generated from the pinned adapter's own
CBOR encoder:

- `wire-join.json`, `wire-peer.json`, `wire-error.json` — the handshake frames.
- `wire-sync.json`, `wire-ephemeral.json` — a framed document message.

## Gateway wiring

The protocol drivers (`ServerConn`, `ClientConn`) are transport-agnostic and
fully tested. Wiring them into the authenticated production WebSocket endpoint
reuses the existing collaboration hub rather than duplicating rooms, persistence,
or cross-instance notification:

- **Auth** — mount the repo route inside the same authenticated router group as
  `/document-versions/{documentVersionID}/sync` and reuse
  `documentCollaborationHandler.authorize`. The repo `PeerId` is never trusted as
  identity.
- **Document authority & persistence** — `hub.acquire` yields a lease over the
  shared `*automerge.Document`; each connection uses its own
  `Document.NewSyncState`. Sync fan-out reuses `lease.NotifyPeers`/`lease.Wake`
  and persistence reuses `lease.SchedulePersist`/`lease.PersistError`, exactly as
  the legacy handler does.
- **Document id** — the frontend chooses the `automerge:<id>` URL, so the gateway
  uses `NewAdoptingServerConn`: it announces nothing on `Start` and binds to the
  id in the client's first `sync`/`request` frame, then answers for that id and
  rejects any other. This removes the need for a server/frontend id-derivation
  contract.
- **Ephemeral (presence/cursors)** — repo presence travels as opaque `ephemeral`
  frames, not the legacy structured snapshots. `ServerConn.Receive` returns a
  non-duplicate ephemeral frame as `fanout`; the handler publishes it with
  `lease.BroadcastEphemeral`, and reads other peers' frames from
  `lease.Ephemeral` to write to its socket. This is scoped to one server
  instance.
- **Selections/carets** — a caret or selection is published as a presence
  `update` whose value is a `TextSelectionValue`: the addressed text field plus a
  stable Automerge anchor and head cursor (the bytes from `Text.Cursor`), never
  integer offsets. Offsets drift when anyone types before the caret; a cursor
  resolves (via `Text.CursorPosition`) to the same character after arbitrary
  concurrent edits, so remote carets stay anchored. The presence layer only
  transports the cursor bytes, keeping it independent of the CRDT engine; the
  server and Go agents create and resolve the cursors.

### Remaining contract decisions before enabling the endpoint

These need the migrated frontend (and a Postgres-backed integration test) to
settle, so they are intentionally not encoded as untested production code yet:

- **Cross-instance ephemeral** — `realtime.Events` currently carries only the
  document-version id (a "changed" signal) over `NOTIFY`. Repo ephemeral gossip
  across server instances needs the payload carried too, or a separate pub/sub
  channel. Single-instance fan-out (above) is complete.
- **Seeding** — the legacy handshake ships `SeedContent` for the client to apply;
  the repo protocol has no such field. The repo endpoint must instead seed the
  server-side document (authoritative) before serving, or rely on the first
  writer. Which of these the frontend expects is undecided.
- **Auth token transport** — a repo client sets no cookies by default; how the
  frontend presents the session/bearer credential on the WebSocket upgrade must
  match `authn` middleware expectations.

## Deliberately deferred

- `remote-subscription-change` and `remote-heads-changed` handling (not needed by
  a single-authority gateway; revisit if multi-storage subscription is wanted).
- Storage adapters: PostgreSQL remains the document authority; we do not adopt
  repo storage.
