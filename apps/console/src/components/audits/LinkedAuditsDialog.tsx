import { getAuditStateVariant } from "@probo/helpers";
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
  LinkedAuditsDialogFragment$data,
  LinkedAuditsDialogFragment$key,
} from "#/__generated__/core/LinkedAuditsDialogFragment.graphql";
import type { LinkedAuditsDialogQuery } from "#/__generated__/core/LinkedAuditsDialogQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

const auditsQuery = graphql`
  query LinkedAuditsDialogQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        ...LinkedAuditsDialogFragment
      }
    }
  }
`;

const auditsFragment = graphql`
  fragment LinkedAuditsDialogFragment on Organization
  @refetchable(queryName: "LinkedAuditsDialogQuery_fragment")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    order: { type: "AuditOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    audits(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "LinkedAuditsDialogQuery_audits") {
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
  linkedAudits?: { id: string }[];
  onLink: (auditId: string) => void;
  onUnlink: (auditId: string) => void;
};

export function LinkedAuditsDialog({ children, ...props }: Props) {
  const { __ } = useTranslate();

  return (
    <Dialog trigger={children} title={__("Link audits")}>
      <DialogContent>
        <Suspense fallback={<Spinner centered />}>
          <LinkedAuditsDialogContent {...props} />
        </Suspense>
      </DialogContent>
      <DialogFooter exitLabel={__("Close")} />
    </Dialog>
  );
}

function LinkedAuditsDialogContent(props: Omit<Props, "children">) {
  const organizationId = useOrganizationId();
  const query = useLazyLoadQuery<LinkedAuditsDialogQuery>(auditsQuery, {
    organizationId,
  });
  const { data, loadNext, hasNext, isLoadingNext } = usePaginationFragment(
    auditsFragment,
    query.organization as LinkedAuditsDialogFragment$key,
  );
  const { __ } = useTranslate();
  const [search, setSearch] = useState("");
  const audits = useMemo(
    () => data.audits?.edges?.map(edge => edge.node) ?? [],
    [data.audits],
  );
  const linkedIds = useMemo(() => {
    return new Set(props.linkedAudits?.map(a => a.id) ?? []);
  }, [props.linkedAudits]);

  const filteredAudits = useMemo(() => {
    return audits.filter(audit =>
      (audit.name || "").toLowerCase().includes(search.toLowerCase()),
    );
  }, [audits, search]);

  return (
    <>
      <div className="flex items-center gap-2 sticky top-0 relative py-4 bg-linear-to-b from-50% from-level-2 to-level-2/0 px-6">
        <Input
          icon={IconMagnifyingGlass}
          placeholder={__("Search audits...")}
          onValueChange={setSearch}
        />
      </div>
      <div className="divide-y divide-border-low">
        {filteredAudits.map(audit => (
          <AuditRow
            key={audit.id}
            audit={audit}
            linkedAudits={linkedIds}
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

type Audit = NodeOf<LinkedAuditsDialogFragment$data["audits"]>;

type RowProps = {
  audit: Audit;
  linkedAudits: Set<string>;
  disabled?: boolean;
  onLink: (auditId: string) => void;
  onUnlink: (auditId: string) => void;
};

function AuditRow(props: RowProps) {
  const { __ } = useTranslate();

  const isLinked = props.linkedAudits.has(props.audit.id);
  const onClick = isLinked ? props.onUnlink : props.onLink;
  const IconComponent = isLinked ? IconTrashCan : IconPlusLarge;

  return (
    <button
      className="py-4 flex items-center gap-4 hover:bg-subtle cursor-pointer px-6 w-full h-[100px]"
      onClick={() => onClick(props.audit.id)}
    >
      <div className="flex flex-col items-start gap-1">
        <div className="font-medium">{props.audit.framework?.name}</div>
        {props.audit.name && (
          <div className="text-sm text-txt-secondary">{props.audit.name}</div>
        )}
      </div>
      <Badge color={getAuditStateVariant(props.audit.state)}>
        {props.audit.state.replace(/_/g, " ")}
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
