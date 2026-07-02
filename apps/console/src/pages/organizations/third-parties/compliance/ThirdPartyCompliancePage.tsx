// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
import {
  Button,
  IconPlusLarge,
  PageHeader,
  Tbody,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import type { ComponentProps } from "react";
import { graphql, type PreloadedQuery, usePreloadedQuery, useRefetchableFragment } from "react-relay";

import type { ThirdPartyCompliancePageFragment$key } from "#/__generated__/core/ThirdPartyCompliancePageFragment.graphql";
import type { ThirdPartyCompliancePageQuery } from "#/__generated__/core/ThirdPartyCompliancePageQuery.graphql";
import type { ThirdPartyCompliancePageRefetchQuery } from "#/__generated__/core/ThirdPartyCompliancePageRefetchQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";

import { UploadComplianceReportDialog } from "../_components/UploadComplianceReportDialog";

import { ThirdPartyComplianceReportRow } from "./_components/ThirdPartyComplianceReportRow";

const complianceReportsFragment = graphql`
  fragment ThirdPartyCompliancePageFragment on ThirdParty
  @refetchable(queryName: "ThirdPartyCompliancePageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "ThirdPartyComplianceReportOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    id
    name
    canUploadComplianceReport: permission(
      action: "core:thirdParty-compliance-report:upload"
    )
    complianceReports(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "ThirdPartyCompliancePage_complianceReports") {
      __id
      edges {
        node {
          id
          ...ThirdPartyComplianceReportRow_report
        }
      }
    }
  }
`;

export const thirdPartyCompliancePageQuery = graphql`
  query ThirdPartyCompliancePageQuery($thirdPartyId: ID!) {
    node(id: $thirdPartyId) {
      __typename
      ... on ThirdParty {
        ...ThirdPartyCompliancePageFragment
      }
    }
  }
`;

interface ThirdPartyCompliancePageProps {
  queryRef: PreloadedQuery<ThirdPartyCompliancePageQuery>;
}

export default function ThirdPartyCompliancePage(props: ThirdPartyCompliancePageProps) {
  const queryData = usePreloadedQuery<ThirdPartyCompliancePageQuery>(thirdPartyCompliancePageQuery, props.queryRef);
  if (queryData.node?.__typename !== "ThirdParty") {
    throw new Error("Third party not found");
  }

  const [data, refetch] = useRefetchableFragment<
    ThirdPartyCompliancePageRefetchQuery,
    ThirdPartyCompliancePageFragment$key
  >(complianceReportsFragment, queryData.node);

  const connectionId = data.complianceReports.__id;
  const reports = data.complianceReports.edges.map(edge => edge.node);
  const { __ } = useTranslate();

  usePageTitle(data.name + " - " + __("Compliance reports"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Compliance reports")}
        description={__("Track third party compliance certifications and reports.")}
      >
        {data.canUploadComplianceReport && (
          <UploadComplianceReportDialog
            thirdPartyId={data.id}
            connectionId={connectionId}
          >
            <Button icon={IconPlusLarge}>{__("Add report")}</Button>
          </UploadComplianceReportDialog>
        )}
      </PageHeader>

      <SortableTable
        refetch={refetch as ComponentProps<typeof SortableTable>["refetch"]}
      >
        <Thead>
          <Tr>
            <Th>{__("Report name")}</Th>
            <SortableTh field="REPORT_DATE">{__("Report date")}</SortableTh>
            <Th>{__("Valid until")}</Th>
            <Th>{__("File size")}</Th>
            {reports.length > 0 && <Th>{__("Actions")}</Th>}
          </Tr>
        </Thead>
        <Tbody>
          {reports.map(report => (
            <ThirdPartyComplianceReportRow
              key={report.id}
              reportKey={report}
              connectionId={connectionId}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}
