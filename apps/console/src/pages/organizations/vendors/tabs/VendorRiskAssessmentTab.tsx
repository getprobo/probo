import { formatDate } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  IconPlusLarge,
  RiskBadge,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  TrButton,
} from "@probo/ui";
import { clsx } from "clsx";
import { type ComponentProps, useState } from "react";
import { useFragment, useRefetchableFragment } from "react-relay";
import { useOutletContext, useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { VendorGraphNodeQuery$data } from "#/__generated__/core/VendorGraphNodeQuery.graphql";
import type { VendorRiskAssessmentTabFragment$key } from "#/__generated__/core/VendorRiskAssessmentTabFragment.graphql";
import type { VendorRiskAssessmentTabFragment_assessment$key } from "#/__generated__/core/VendorRiskAssessmentTabFragment_assessment.graphql";
import type { VendorRiskAssessmentTabQuery } from "#/__generated__/core/VendorRiskAssessmentTabQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";

import { CreateRiskAssessmentDialog } from "../dialogs/CreateRiskAssessmentDialog";

const riskAssessmentsFragment = graphql`
  fragment VendorRiskAssessmentTabFragment on Vendor
  @refetchable(queryName: "VendorRiskAssessmentTabQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "VendorRiskAssessmentOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    id

    riskAssessments(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "VendorRiskAssessmentTabFragment_riskAssessments") {
      __id
      edges {
        node {
          id
          ...VendorRiskAssessmentTabFragment_assessment
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

const riskAssessmentFragment = graphql`
  fragment VendorRiskAssessmentTabFragment_assessment on VendorRiskAssessment {
    id
    createdAt
    expiresAt
    dataSensitivity
    businessImpact
    notes
  }
`;

export default function VendorRiskAssessmentTab() {
  const { vendor } = useOutletContext<{
    vendor: VendorGraphNodeQuery$data["node"];
  }>();
  const [data, refetch] = useRefetchableFragment<
    VendorRiskAssessmentTabQuery,
    VendorRiskAssessmentTabFragment$key
  >(riskAssessmentsFragment, vendor);
  const assessments = data.riskAssessments.edges.map(edge => edge.node);
  const { __ } = useTranslate();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);
  const [expanded, setExpanded] = useState<string | null>(null);

  usePageTitle(vendor.name + " - " + __("Risk Assessments"));

  if (assessments.length === 0) {
    return (
      <div className="text-center text-sm py-6 text-txt-secondary flex flex-col items-center gap-2">
        {__("No risk assessments found")}
        {!isSnapshotMode && vendor.canCreateRiskAssessment && (
          <CreateRiskAssessmentDialog
            vendorId={vendor.id}
            connection={data.riskAssessments.__id}
          >
            <Button icon={IconPlusLarge} variant="secondary">
              {__("Add Risk Assessment")}
            </Button>
          </CreateRiskAssessmentDialog>
        )}
      </div>
    );
  }

  return (
    <div className="space-y-6 relative">
      <div className="flex justify-end"></div>
      <div className="overflow-x-auto">
        <SortableTable
          refetch={refetch as ComponentProps<typeof SortableTable>["refetch"]}
        >
          <Thead>
            <Tr>
              <SortableTh field="CREATED_AT">{__("Created At")}</SortableTh>
              <SortableTh field="EXPIRES_AT">{__("Expires")}</SortableTh>
              <Th>{__("Data sensitivity")}</Th>
              <Th>{__("Business impact")}</Th>
            </Tr>
          </Thead>
          <Tbody>
            {!isSnapshotMode && vendor.canCreateRiskAssessment && (
              <CreateRiskAssessmentDialog
                vendorId={vendor.id}
                connection={data.riskAssessments.__id}
              >
                <TrButton colspan={5} onClick={() => {}}>
                  {__("Add Risk Assessment")}
                </TrButton>
              </CreateRiskAssessmentDialog>
            )}
            {assessments.map(assessment => (
              <AssessmentRow
                key={assessment.id}
                assessmentKey={assessment}
                isExpanded={expanded === assessment.id}
                onClick={() =>
                  setExpanded(prev =>
                    prev === assessment.id ? null : assessment.id,
                  )}
              />
            ))}
          </Tbody>
        </SortableTable>
      </div>
    </div>
  );
}

type AssessmentRowProps = {
  assessmentKey: VendorRiskAssessmentTabFragment_assessment$key;
  onClick: (id: string) => void;
  isExpanded: boolean;
};

function AssessmentRow(props: AssessmentRowProps) {
  const { __ } = useTranslate();
  const assessment
    = useFragment<VendorRiskAssessmentTabFragment_assessment$key>(
      riskAssessmentFragment,
      props.assessmentKey,
    );
  const { relativeDateFormat } = useTranslate();
  const isExpired = new Date(assessment.expiresAt) < new Date();

  return (
    <>
      <Tr
        className={clsx(
          isExpired && "opacity-50",
          "cursor-pointer",
          props.isExpanded && "border-none",
        )}
        onClick={() => props.onClick(assessment.id)}
      >
        <Td>
          <span className="text-xs text-txt-secondary ml-1">
            {formatDate(assessment.createdAt)}
          </span>
        </Td>
        <Td>
          <div className="flex items-center gap-2">
            {relativeDateFormat(assessment.expiresAt)}
            {isExpired && <Badge variant="neutral">{__("Expired")}</Badge>}
          </div>
        </Td>
        <Td>
          <RiskBadge level={assessment.dataSensitivity} />
        </Td>
        <Td>
          <RiskBadge level={assessment.businessImpact} />
        </Td>
      </Tr>
      {props.isExpanded && (
        <Tr className={clsx("border-none", isExpired && "opacity-50")}>
          <Td colSpan={4}>
            <div className="space-y-2">
              <div>
                {__("Notes")}
                :
              </div>
              <p className="text-txt-secondary">{assessment.notes}</p>
            </div>
          </Td>
        </Tr>
      )}
    </>
  );
}
