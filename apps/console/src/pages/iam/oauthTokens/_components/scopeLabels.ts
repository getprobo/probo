// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

const apiScopeLabels: Record<string, string> = {
  "v1:access-review:read": "Read access reviews",
  "v1:access-review": "Manage access reviews",
  "v1:agent:read": "Read agents",
  "v1:agent": "Manage agents",
  "v1:asset:read": "Read assets",
  "v1:asset": "Manage assets",
  "v1:audit:read": "Read audits",
  "v1:audit": "Manage audits",
  "v1:common-third-party:read": "Read common third parties",
  "v1:common-third-party": "Manage common third parties",
  "v1:compliance-page:read": "Read compliance pages",
  "v1:compliance-page": "Manage compliance pages",
  "v1:connector:read": "Read connectors",
  "v1:connector": "Manage connectors",
  "v1:control:read": "Read controls",
  "v1:control": "Manage controls",
  "v1:datum:read": "Read data",
  "v1:datum": "Manage data",
  "v1:document:read": "Read documents",
  "v1:document": "Manage documents",
  "v1:iam:read": "Read IAM settings",
  "v1:iam": "Manage IAM settings",
  "v1:org:read": "Read organization",
  "v1:org": "Manage organization",
  "v1:privacy:read": "Read privacy settings",
  "v1:privacy": "Manage privacy settings",
  "v1:risk:read": "Read risks",
  "v1:risk": "Manage risks",
  "v1:slack-connection:read": "Read Slack connections",
  "v1:slack-connection": "Manage Slack connections",
  "v1:task:read": "Read tasks",
  "v1:task": "Manage tasks",
  "v1:third-party:read": "Read third parties",
  "v1:third-party": "Manage third parties",
  "v1:webhook:read": "Read webhooks",
  "v1:webhook": "Manage webhooks",
};

export function formatApiScopeLabel(scope: string): string {
  return apiScopeLabels[scope] ?? scope;
}
