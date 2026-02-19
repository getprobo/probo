import { getReportStateVariant, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  Card,
  IconChevronDown,
  IconPlusLarge,
  IconTrashCan,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  TrButton,
} from "@probo/ui";
import { clsx } from "clsx";
import { useMemo, useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { LinkedReportsCardFragment$key } from "#/__generated__/core/LinkedReportsCardFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { LinkedReportsDialog } from "./LinkedReportsDialog";

const linkedReportFragment = graphql`
  fragment LinkedReportsCardFragment on Report {
    id
    name
    createdAt
    state
    validFrom
    validUntil
    framework {
      id
      name
    }
  }
`;

type Mutation<Params> = (p: {
  variables: {
    input: {
      reportId: string;
    } & Params;
    connections: string[];
  };
}) => void;

type Props<Params> = {
  reports: (LinkedReportsCardFragment$key & { id: string })[];
  params: Params;
  disabled?: boolean;
  connectionId: string;
  onAttach: Mutation<Params>;
  onDetach: Mutation<Params>;
  variant?: "card" | "table";
  readOnly?: boolean;
};

export function LinkedReportsCard<Params>(props: Props<Params>) {
  const { __ } = useTranslate();
  const [limit, setLimit] = useState<number | null>(4);
  const reports = useMemo(() => {
    return limit ? props.reports.slice(0, limit) : props.reports;
  }, [props.reports, limit]);
  const showMoreButton = limit !== null && props.reports.length > limit;
  const variant = props.variant ?? "table";

  const onAttach = (reportId: string) => {
    props.onAttach({
      variables: {
        input: {
          reportId,
          ...props.params,
        },
        connections: [props.connectionId],
      },
    });
  };

  const onDetach = (reportId: string) => {
    props.onDetach({
      variables: {
        input: {
          reportId,
          ...props.params,
        },
        connections: [props.connectionId],
      },
    });
  };

  const Wrapper = variant === "card" ? Card : "div";

  return (
    <Wrapper padded className="space-y-[10px]">
      {variant === "card" && (
        <div className="flex justify-between">
          <div className="text-lg font-semibold">{__("Reports")}</div>
          {!props.readOnly && (
            <LinkedReportsDialog
              disabled={props.disabled}
              linkedReports={props.reports}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <Button variant="tertiary" icon={IconPlusLarge}>
                {__("Link report")}
              </Button>
            </LinkedReportsDialog>
          )}
        </div>
      )}
      <Table className={clsx(variant === "card" && "bg-invert")}>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("State")}</Th>
            {!props.readOnly && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {reports.length === 0 && (
            <Tr>
              <Td
                colSpan={props.readOnly ? 2 : 3}
                className="text-center text-txt-secondary"
              >
                {__("No reports linked")}
              </Td>
            </Tr>
          )}
          {reports.map(report => (
            <ReportRow
              key={report.id}
              report={report}
              onClick={onDetach}
              readOnly={props.readOnly}
            />
          ))}
          {variant === "table" && !props.readOnly && (
            <LinkedReportsDialog
              disabled={props.disabled}
              linkedReports={props.reports}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <TrButton colspan={3} icon={IconPlusLarge}>
                {__("Link report")}
              </TrButton>
            </LinkedReportsDialog>
          )}
        </Tbody>
      </Table>
      {showMoreButton && (
        <Button
          variant="tertiary"
          onClick={() => setLimit(null)}
          className="mt-3 mx-auto"
          icon={IconChevronDown}
        >
          {sprintf(__("Show %s more"), props.reports.length - limit)}
        </Button>
      )}
    </Wrapper>
  );
}

function ReportRow(props: {
  report: LinkedReportsCardFragment$key & { id: string };
  onClick: (reportId: string) => void;
  readOnly?: boolean;
}) {
  const report = useFragment(linkedReportFragment, props.report);
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  return (
    <Tr to={`/organizations/${organizationId}/reports/${report.id}`}>
      <Td>
        <div className="flex flex-col">
          <div className="font-medium">{report.framework?.name}</div>
          {report.name && (
            <div className="text-sm text-txt-secondary">{report.name}</div>
          )}
        </div>
      </Td>
      <Td>
        <Badge color={getReportStateVariant(report.state)}>
          {report.state.replace(/_/g, " ")}
        </Badge>
      </Td>
      {!props.readOnly && (
        <Td noLink width={50} className="text-end">
          <Button
            variant="secondary"
            onClick={() => props.onClick(report.id)}
            icon={IconTrashCan}
          >
            {__("Unlink")}
          </Button>
        </Td>
      )}
    </Tr>
  );
}
