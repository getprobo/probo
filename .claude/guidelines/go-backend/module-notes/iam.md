# Probo — Go Backend — pkg/iam (and sub-packages)

**Purpose.** Identity, authentication, and authorization for the entire
backend. `pkg/iam/policy` is the pure value-object layer (Policy /
Statement / Condition / Evaluator); `pkg/iam` is the orchestrator
(`Authorizer`, `PolicySet`, action constants); `pkg/iam/oidc`,
`pkg/iam/saml`, `pkg/iam/scim`, `pkg/iam/oauth2server` are identity-provider
integrations.

> See [patterns.md § 4](../patterns.md#4-authorization).

**Key files.**

- `pkg/iam/policy/policy.go`, `statement.go`, `evaluator.go` — fluent
  policy DSL. Decision order: explicit Deny > explicit Allow >
  implicit deny.
- `pkg/iam/iam_actions.go` — IAM action constants (`iam:organization:*`).
- `pkg/iam/iam_policy_set.go` — IAM policies (admin/owner/employee/...).
- `pkg/iam/authorizer.go` — `Authorizer` (loads attributes, membership,
  selects policies, evaluates) + `AuthorizationAttributer` interface
  (entities implement it to expose attributes).
- `pkg/server/api/authz/authorization.go` — `AuthorizeFunc` factory
  used by GraphQL resolvers.

**How to extend (a new product action).**

1. Add the action constant to `pkg/probo/actions.go` (or
   `pkg/iam/iam_actions.go` for IAM-domain actions).
2. Add the `Allow` statement to `pkg/probo/policies.go` — every product
   policy **must** include `organizationCondition()` to scope it to
   the principal's org. Without it, the policy leaks across tenants.
3. Call `r.authorize(ctx, resourceID, probo.ActionXxx)` as the first
   line of every resolver method that uses the action.
4. Verify with an RBAC matrix e2e test (see
   [testing.md § 4](../testing.md#4-test-patterns)).

**Top pitfalls.**

- Skipping `organizationCondition` — over-permissive cross-tenant policy.
- Following `contrib/claude/authorization.md` literally and using
  `policy.In(...)` — the helper does not exist. Use the `Condition`
  struct directly with `ConditionOperator.In`. See
  [pitfalls.md § 2](../pitfalls.md).
- Logging the OIDC `error_description` verbatim — PII leak. See
  [pitfalls.md § 3](../pitfalls.md).
- Failed PKCE that does not delete the auth code — replay attack
  surface (PR #957). See [pitfalls.md § 20](../pitfalls.md).
- 100% unit-test coverage requirement on auth-sensitive code is
  enforced in review.
