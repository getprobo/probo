// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, Card } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Link } from "react-router";
import { graphql } from "relay-runtime";

import type { EmployeeDevicesPageQuery } from "#/__generated__/core/EmployeeDevicesPageQuery.graphql";

import { CreateEnrollmentTokenForm } from "./_components/CreateEnrollmentTokenForm";

export const employeeDevicesPageQuery = graphql`
  query EmployeeDevicesPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canCreateEnrollmentToken: permission(
          action: "core:device-enrollment-token:create"
        )
      }
    }
  }
`;

interface EmployeeDevicesPageProps {
  queryRef: PreloadedQuery<EmployeeDevicesPageQuery>;
}

export function EmployeeDevicesPage({ queryRef }: EmployeeDevicesPageProps) {
  const { __ } = useTranslate();

  usePageTitle(__("Devices"));

  const { organization } = usePreloadedQuery(
    employeeDevicesPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }

  return (
    <div className="space-y-6">
      <header className="space-y-1">
        <h1 className="text-2xl font-semibold">{__("Enroll your device")}</h1>
        <p className="text-sm text-tertiary">
          {__(
            "Use browser enrollment for the easiest setup. You can still generate tokens below for manual CLI or MDM installs.",
          )}
        </p>
      </header>

      {organization.canCreateEnrollmentToken && (
        <Card padded className="space-y-2">
          <h2 className="text-base font-medium">{__("Recommended")}</h2>
          <p className="text-sm text-tertiary">
            {__(
              "Open browser enrollment, choose your organization, then click the enroll button to open the Probo agent.",
            )}
          </p>
          <Button asChild>
            <Link to="/enroll">{__("Open browser enrollment")}</Link>
          </Button>
        </Card>
      )}

      {organization.canCreateEnrollmentToken && <CreateEnrollmentTokenForm />}
    </div>
  );
}
