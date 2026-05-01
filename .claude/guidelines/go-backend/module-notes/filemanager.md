# Probo — Go Backend — pkg/filemanager + pkg/filevalidation

**Purpose.** S3-compatible blob storage abstraction (used against
SeaweedFS in dev, S3 in prod) plus MIME-type validation guards.

**Key files.**

- `pkg/filemanager/service.go` — `Service`, `Upload`, `Get`,
  `GetFileBase64`, `Delete`, presigned URLs.
- `pkg/filevalidation/validate.go` — MIME-type allow-lists per use case
  (evidence, document import, logo upload).

**How to use.**

- Upload: stream the file body into `Upload`, get back a storage key.
- Read: `Get` for `io.ReadCloser` streaming;
  `GetFileBase64` when the consumer needs it as a base64 string (e.g.
  the LLM `FilePart` payload — see
  [evidencedescriber pattern](./llm.md)).

**Top pitfalls.**

- Using a presigned URL with too long a TTL — keep TTLs minutes, not
  hours.
- Skipping `pkg/filevalidation` on user-supplied uploads — gives users
  a free path to upload arbitrary content. Always validate MIME against
  the allow-list relevant to the use case.
- Holding open file handles across goroutine boundaries — close before
  returning to the worker loop.
