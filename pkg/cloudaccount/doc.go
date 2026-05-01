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

// Package cloudaccount provides the multi-cloud connector abstraction
// powering Probo's cloud-account feature: AWS via cross-account
// AssumeRole, GCP via service-account JSON keys, and Azure via App
// Registration client secrets. Each provider type exposes a typed
// SDK surface (e.g. *AWSProvider.IAM(), *GCPProvider.CRM(),
// *AzureProvider.ARM()) so downstream consumers (the access-review
// drivers, the periodic probe worker) get static types and never
// need runtime type assertions.
//
// # Architectural rules
//
// This package operates on a thin CloudAccountRecord value type
// declared in credentials.go. It DOES NOT import pkg/coredata's
// *CloudAccount entity -- the service layer (pkg/probo) maps a
// coredata entity to a CloudAccountRecord before calling the
// registry. The package legitimately imports the four pure-enum
// types (CloudAccountProvider, CloudAccountCredentialKind,
// CloudAccountScopeKind, CloudAccountAuditModule) from coredata
// because they are pure value types with no I/O.
//
// # Credentials envelope
//
// All persisted credential payloads are JSON envelopes of the
// shape {"v":1,"kind":"...","payload":{...}}. The kind inside the
// envelope is the source of truth; UnmarshalCredentials rejects
// envelopes whose kind is incompatible with the expected provider.
//
// # Probe lifecycle
//
// Each typed provider implements the Probeable interface via its
// own Probe(ctx) method. The probe worker uses Registry.BuildProbeable
// (a polymorphic dispatch) to obtain the Probeable for a given
// record without caring which concrete provider type backs it.
// Drivers, by contrast, take concrete types (e.g. *AWSProvider) so
// type-checked compile-time dispatch is preserved.
//
// # Logging hygiene
//
// Code in this package and all callers MUST NOT log: external_id,
// role ARNs, GCP service-account emails, JSON keys, Azure
// tenant/client/secret values, or scope identifiers. The persisted
// CloudAccount.LastProbeError column is the operator-facing
// diagnostic surface; duplicating it in logs both inflates exposure
// and creates a divergence risk when the column is later redacted.
package cloudaccount
