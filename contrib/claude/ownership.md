# Resource ownership

**Status: provisional.** This documents the pattern the codebase follows today. The
team has not formally ratified profile-based ownership over identity-based
ownership; treat this guide as the default for new work until that decision is
revisited.

## Model

Organization-scoped resources name a **membership profile** as owner — the
person's membership in that organization — not the global **identity**.

| Concept | GID entity type | Scope |
| ------- | --------------- | ----- |
| Identity | `Identity` | Global person (login, sessions) |
| Membership profile | `MembershipProfile` | Person within one organization |

Ownership answers: *who in this org is responsible for this resource?* That is
always a membership profile, even when the UX speaks in terms of "people".

## Layer conventions

Use the same shape across DB, services, GraphQL, and console pickers.

| Layer | Convention |
| ----- | ------------ |
| Database column | `owner_profile_id TEXT REFERENCES iam_membership_profiles(id) ON UPDATE CASCADE ON DELETE RESTRICT` |
| Go struct field | `OwnerID` with `` `db:"owner_profile_id"` `` (name stays `OwnerID`; tag names the column) |
| GraphQL input | `ownerId: ID` on create/update mutations |
| GraphQL output | `owner: Profile` resolved via profile dataloader |
| Validation | `validator.GID(coredata.MembershipProfileEntityType)` |
| Service create/update | `profile.LoadByID` inside the transaction; verify profile belongs to the resource's `organization_id` when not already enforced by scope. Self-service enroll accepts `IdentityID` and uses `LoadByIdentityIDAndOrganizationID`; admin create uses explicit `OwnerID` only |
| Console people picker | [`PeopleSelectField`](../../apps/console/src/components/form/PeopleSelectField.tsx) — value is `Profile.id` |

### Resources that follow this pattern

- Assets, data (datum), risks, obligations, findings
- Third parties (`business_owner_profile_id`, `security_owner_profile_id`)
- Devices (ITAM) — aligned with compliance resources as of the devices table
  introduction

## GraphQL resolver pattern

Store only the profile GID on the GraphQL type; resolve `owner` in a field
resolver. Authorize the **profile** GID only — do not re-authorize the parent
resource for the `owner` field (parent access is already established).

Default for new work is a **nullable** owner (`OwnerID *gid.GID`, GraphQL
`owner: Profile`). Set the embedded profile only when present so the resolver
nil guard is live (device, risk, finding, third-party owners):

```go
// types — set only when present
if resource.OwnerID != nil {
	obj.Owner = &Profile{ID: *resource.OwnerID}
}

// resolver — authorize profile, then load
if obj.Owner == nil {
	return nil, nil
}

if _, err := r.authorize(ctx, obj.Owner.ID, iam.ActionMembershipProfileGet); err != nil {
	return nil, err
}

owner, err := loaders.Profile.Load(ctx, obj.Owner.ID)
```

For a **required** owner (`OwnerID gid.GID`, GraphQL `owner: Profile!` — asset,
datum, obligation), always embed `Owner: &Profile{ID: …}` in the constructor and
omit the `obj.Owner == nil` guard.

Do not reverse-lookup identity → profile at read time when the profile id is
already stored.

## Employee and self-service flows

Authenticated users are **identities**. Device create vs enroll:

- **`enrollDevice`** (employee self-service): resolver passes `identity.ID`;
  service resolves the membership profile with
  `LoadByIdentityIDAndOrganizationID` and always sets that profile as owner.
- **`createDevice`** (admin): resolver passes only the optional `ownerId` from
  the picker. No identity default — omitted/`null` means unowned.

List/count filters that key on `owner_profile_id` (e.g. `viewer.enrolledDevices`)
may still resolve identity → profile at the resolver until a list-by-identity
helper exists.

## IAM when policies compare identity

Some policies match `principal.id` (identity) to a resource attribute, e.g.
ITAM employee access to own devices:

```go
ownerCondition = policy.Equals("principal.id", "resource.owner_id")
```

Storage is profile-based, but the policy still compares identities. Bridge in
`AuthorizationAttributes` with two entity queries — no cross-entity JOINs (see
[`coredata.md`](coredata.md#no-cross-entity-joins)):

```go
// Step 1 — devices table only
SELECT id, organization_id, owner_profile_id
FROM devices
WHERE id = ANY(@resource_ids::text[])

// Step 2 — profiles table only (e.g. MembershipProfile.AuthorizationAttributes)
SELECT id, identity_id
FROM iam_membership_profiles
WHERE id = ANY(@profile_ids::text[])

// Step 3 — map in Go: attrs["owner_id"] = identityByProfileID[ownerProfileID]
```

Policy code stays unchanged; only the attributer translates profile storage →
identity comparison.

## Checklist for a new owned resource

1. Migration: `owner_profile_id` FK to `iam_membership_profiles`.
2. Entity struct: `OwnerID` with `owner_profile_id` db tag.
3. GraphQL: `ownerId` input, `owner: Profile` output with `forceResolver`.
4. Service: validate `MembershipProfileEntityType`; load profile in tx. For
   self-service enroll, accept `IdentityID` and resolve with
   `LoadByIdentityIDAndOrganizationID`. Admin create uses explicit `OwnerID`
   only (nil means unowned).
5. Types: for nullable owners, set `Owner` only when `OwnerID != nil`; for
   required owners (`Profile!`), always embed `Owner: &Profile{ID: …}`.
6. Resolver `owner` field: nil-guard when nullable; `authorize` with
   `ActionMembershipProfileGet` on the profile GID, then `loaders.Profile.Load`
   (no parent-resource authorize).
7. If employee self-service lists by caller: resolve identity → profile (resolver
   or service) and query by `owner_profile_id`.
8. If IAM compares `principal.id` to owner: bridge identity in
   `AuthorizationAttributes` (see above).

## Open question (team review)

An alternative is storing **identity** GIDs for person-centric resources (e.g.
devices tied to a person across org context) and accepting profile ids at the
API boundary via normalization. Compliance resources migrated from `peoples` to
membership profiles in [`20260203T132700Z.sql`](../../pkg/coredata/migrations/20260203T132700Z.sql);
devices were added on the profile model to stay consistent.

When adding ownership to a new resource, default to **profile** unless there is
a documented reason to anchor on identity. Raise identity-based ownership in
design review if the resource lifecycle is person-global rather than
org-scoped.

## Related

- [`coredata.md`](coredata.md) — entity structs, migrations, `AuthorizationAttributes`
- [`authorization.md`](authorization.md) — policy conditions and attributers
- [`validation.md`](validation.md) — `validator.GID` entity-type checks
- [`graphql.md`](graphql.md) — `@goField(forceResolver: true)` for `owner`
