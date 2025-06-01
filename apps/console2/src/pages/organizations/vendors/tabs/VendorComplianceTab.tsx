import { useOutletContext } from "react-router";
import { graphql } from "relay-runtime";
import type { VendorComplianceTabFragment$key } from "./__generated__/VendorComplianceTabFragment.graphql";
import { useTranslate } from "@probo/i18n";
import { usePageTitle } from "@probo/hooks";

const complianceReportsFragment = graphql`
  fragment VendorComplianceTabFragment on Vendor
  @refetchable(queryName: "ComplianceReportListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "VendorComplianceReportOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    complianceReports(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "VendorComplianceTabFragment_complianceReports") {
      edges {
        node {
          ...VendorComplianceTabFragment_report
        }
      }
    }
  }
`;

const complianceReportFragment = graphql`
  fragment VendorComplianceTabFragment_report on VendorComplianceReport {
    reportDate
    validUntil
    reportName
    fileUrl
    fileSize
  }
`;

export default function VendorComplianceTab() {
  const { vendor } = useOutletContext<{
    vendor: VendorComplianceTabFragment$key & { name: string };
  }>();
  const { __ } = useTranslate();

  usePageTitle(vendor.name + " - " + __("Compliance reports"));
  return <div>Vendor Compliance</div>;
}
