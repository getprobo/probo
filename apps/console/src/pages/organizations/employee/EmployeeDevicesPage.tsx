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
import { Card, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { EmployeeDevicesPageQuery } from "#/__generated__/core/EmployeeDevicesPageQuery.graphql";

import { CreateEnrollmentTokenForm } from "./_components/CreateEnrollmentTokenForm";
import { EnrollmentTokenRow } from "./_components/EnrollmentTokenRow";

export const employeeDevicesPageQuery = graphql`
  query EmployeeDevicesPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        id
        canCreateEnrollmentToken: permission(
          action: "core:device-enrollment-token:create"
        )
        deviceEnrollmentTokens(first: 100)
          @connection(
            key: "EmployeeDevicesPage_deviceEnrollmentTokens"
            filters: []
          ) {
          __id
          edges {
            node {
              id
              ...EnrollmentTokenRowFragment
            }
          }
        }
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

  const tokens = organization.deviceEnrollmentTokens.edges.map(e => e.node);
  const connectionId = organization.deviceEnrollmentTokens.__id;

  return (
    <div className="space-y-6">
      <header className="space-y-1">
        <h1 className="text-2xl font-semibold">{__("Enroll your device")}</h1>
        <p className="text-sm text-tertiary">
          {__(
            "Run the Probo agent on your computer to verify it meets your organization's posture requirements. Generate an enrollment token below and follow the install instructions on the device.",
          )}
        </p>
      </header>

      {organization.canCreateEnrollmentToken && (
        <CreateEnrollmentTokenForm connectionId={connectionId} />
      )}

      <section>
        <h2 className="text-lg font-medium mb-2">
          {__("Active enrollment tokens")}
        </h2>
        <Card>
          <table className="w-full text-sm">
            <Thead>
              <Tr>
                <Th>{__("Name")}</Th>
                <Th>{__("Created at")}</Th>
                <Th>{__("Expires at")}</Th>
                <Th>{__("Usage")}</Th>
                <Th></Th>
              </Tr>
            </Thead>
            <Tbody>
              {tokens.length === 0
                ? (
                    <Tr>
                      <Td colSpan={5} className="text-center text-tertiary">
                        {__("No enrollment tokens yet.")}
                      </Td>
                    </Tr>
                  )
                : (
                    tokens.map(token => (
                      <EnrollmentTokenRow key={token.id} fKey={token} />
                    ))
                  )}
            </Tbody>
          </table>
        </Card>
      </section>
    </div>
  );
}
