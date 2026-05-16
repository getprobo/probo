# Probo — Go Backend — pkg/probo

**Purpose.** Core domain service layer. Implements every business-logic
service for every entity (vendor, control, risk, finding, document,
asset, audit, ...) plus cross-tenant export workers. `Service` is
the root, `Service.WithTenant(tenantID)` returns a `*TenantService`,
and each entity sub-service hangs off `TenantService` as a public
field.

> See [patterns.md § 1](../patterns.md#1-service--tenantservice--the-domain-service-shape)
> and [§ 2 Request+Validate](../patterns.md#2-request--validate).

**Key files.**

- `service.go` — `Service`, `TenantService`, `WithTenant`, root
  workers (`ExportJob`, `lockExportJob`).
- `vendor_service.go` — canonical sub-service: full CRUD,
  `Request + Validate`, `pg.WithTx` with in-tx `webhook.InsertData`,
  double-pointer optional update fields.
- `finding_service.go` — cross-field validation (risk_id required when
  status=risk_accepted).
- `document_service.go` — complex sub-service with custom error types
  (`ErrSignatureNotCancellable`, `ErrDocumentVersionNotDraft`, ...) and
  PDF export / signing flows.
- `actions.go` — all `core:*` action constants.
- `policies.go` — role-keyed policies registered into the IAM
  Authorizer at startup.
- `evidence_description_worker.go` — canonical worker (Claim / Process /
  RecoverStale, FOR UPDATE SKIP LOCKED).

**How to extend (a new entity).**

The 4-file checklist (see [pitfalls.md § 11](../pitfalls.md)):

1. `pkg/probo/<entity>_service.go` — sub-service struct with `svc
   *TenantService` field, `Request + Validate`, `pg.WithTx`, scope use.
2. `pkg/probo/service.go` — add `*<Entity>Service` field to
   `TenantService` struct + initialise it in `WithTenant`.
3. `pkg/probo/actions.go` — `Action<Entity><Verb>` constants.
4. `pkg/probo/policies.go` — `Allow` statements per role with
   `organizationCondition`.

Plus the coredata side (entity file + filter + order field +
entity_type_reg).

**Top pitfalls.**

- Performing authorization inside a service method. **Don't.** Service
  methods are auth-free; resolvers authorize first.
- Constructing a fresh `Scoper` inside a sub-service method. Always use
  `s.svc.scope`. Only root-Service workers (e.g. `ExportJob`) build
  their own scope from a claimed entity ID.
- Calling `webhook.InsertData` outside the entity-mutation transaction —
  partial-state risk.
- Using `*string` instead of `**string` for nullable update fields.
