import { getReportStateVariant } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  IconMagnifyingGlass,
  IconPlusLarge,
  IconTrashCan,
  InfiniteScrollTrigger,
  Input,
  Spinner,
} from "@probo/ui";
import { type ReactNode, Suspense, useMemo, useState } from "react";
import { useLazyLoadQuery, usePaginationFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type {
  LinkedReportsDialogFragment$data,
  LinkedReportsDialogFragment$key,
} from "#/__generated__/core/LinkedReportsDialogFragment.graphql";
import type { LinkedReportsDialogQuery } from "#/__generated__/core/LinkedReportsDialogQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

const reportsQuery = graphql`
  query LinkedReportsDialogQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        ...LinkedReportsDialogFragment
      }
    }
  }
`;

const reportsFragment = graphql`
  fragment LinkedReportsDialogFragment on Organization
  @refetchable(queryName: "LinkedReportsDialogQuery_fragment")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    order: { type: "ReportOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    reports(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "LinkedReportsDialogQuery_reports") {
      edges {
        node {
          id
          name
          state
          validFrom
          validUntil
          framework {
            id
            name
          }
        }
      }
    }
  }
`;

type Props = {
  children: ReactNode;
  disabled?: boolean;
  linkedReports?: { id: string }[];
  onLink: (reportId: string) => void;
  onUnlink: (reportId: string) => void;
};

export function LinkedReportsDialog({ children, ...props }: Props) {
  const { __ } = useTranslate();

  return (
    <Dialog trigger={children} title={__("Link reports")}>
      <DialogContent>
        <Suspense fallback={<Spinner centered />}>
          <LinkedReportsDialogContent {...props} />
        </Suspense>
      </DialogContent>
      <DialogFooter exitLabel={__("Close")} />
    </Dialog>
  );
}

function LinkedReportsDialogContent(props: Omit<Props, "children">) {
  const organizationId = useOrganizationId();
  const query = useLazyLoadQuery<LinkedReportsDialogQuery>(reportsQuery, {
    organizationId,
  });
  const { data, loadNext, hasNext, isLoadingNext } = usePaginationFragment(
    reportsFragment,
    query.organization as LinkedReportsDialogFragment$key,
  );
  const { __ } = useTranslate();
  const [search, setSearch] = useState("");
  const reports = useMemo(
    () => data.reports?.edges?.map(edge => edge.node) ?? [],
    [data.reports],
  );
  const linkedIds = useMemo(() => {
    return new Set(props.linkedReports?.map(a => a.id) ?? []);
  }, [props.linkedReports]);

  const filteredReports = useMemo(() => {
    return reports.filter(report =>
      (report.name || "").toLowerCase().includes(search.toLowerCase()),
    );
  }, [reports, search]);

  return (
    <>
      <div className="flex items-center gap-2 sticky top-0 relative py-4 bg-linear-to-b from-50% from-level-2 to-level-2/0 px-6">
        <Input
          icon={IconMagnifyingGlass}
          placeholder={__("Search reports...")}
          onValueChange={setSearch}
        />
      </div>
      <div className="divide-y divide-border-low">
        {filteredReports.map(report => (
          <ReportRow
            key={report.id}
            report={report}
            linkedReports={linkedIds}
            onLink={props.onLink}
            onUnlink={props.onUnlink}
            disabled={props.disabled}
          />
        ))}
        {hasNext && (
          <InfiniteScrollTrigger
            loading={isLoadingNext}
            onView={() => loadNext(20)}
          />
        )}
      </div>
    </>
  );
}

type Report = NodeOf<LinkedReportsDialogFragment$data["reports"]>;

type RowProps = {
  report: Report;
  linkedReports: Set<string>;
  disabled?: boolean;
  onLink: (reportId: string) => void;
  onUnlink: (reportId: string) => void;
};

function ReportRow(props: RowProps) {
  const { __ } = useTranslate();

  const isLinked = props.linkedReports.has(props.report.id);
  const onClick = isLinked ? props.onUnlink : props.onLink;
  const IconComponent = isLinked ? IconTrashCan : IconPlusLarge;

  return (
    <button
      className="py-4 flex items-center gap-4 hover:bg-subtle cursor-pointer px-6 w-full h-[100px]"
      onClick={() => onClick(props.report.id)}
    >
      <div className="flex flex-col items-start gap-1">
        <div className="font-medium">{props.report.framework?.name}</div>
        {props.report.name && (
          <div className="text-sm text-txt-secondary">{props.report.name}</div>
        )}
      </div>
      <Badge color={getReportStateVariant(props.report.state)}>
        {props.report.state.replace(/_/g, " ")}
      </Badge>
      <Button
        disabled={props.disabled}
        className="ml-auto"
        variant={isLinked ? "secondary" : "primary"}
        asChild
      >
        <span>
          <IconComponent size={16} />
          {" "}
          {isLinked ? __("Unlink") : __("Link")}
        </span>
      </Button>
    </button>
  );
}
