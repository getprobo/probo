import {
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
  Badge,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { getObligationStatusVariant, getObligationStatusLabel } from "@probo/helpers";
import { Suspense, useMemo, useState, type ReactNode } from "react";
import { graphql } from "relay-runtime";
import { useLazyLoadQuery, usePaginationFragment } from "react-relay";
import type { LinkedObligationsDialogQuery } from "./__generated__/LinkedObligationsDialogQuery.graphql";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { NodeOf } from "/types";
import type {
  LinkedObligationsDialogFragment$data,
  LinkedObligationsDialogFragment$key,
} from "./__generated__/LinkedObligationsDialogFragment.graphql";

const obligationsQuery = graphql`
  query LinkedObligationsDialogQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        ...LinkedObligationsDialogFragment
      }
    }
  }
`;

const obligationsFragment = graphql`
  fragment LinkedObligationsDialogFragment on Organization
  @refetchable(queryName: "LinkedObligationsDialogQuery_fragment")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    order: { type: "ObligationOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    obligations(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "LinkedObligationsDialogQuery_obligations") {
      edges {
        node {
          id
          requirement
          area
          source
          status
          owner {
            fullName
          }
        }
      }
    }
  }
`;

type Props = {
  children: ReactNode;
  connectionId: string;
  disabled?: boolean;
  linkedObligations?: { id: string }[];
  onLink: (obligationId: string) => void;
  onUnlink: (obligationId: string) => void;
};

export function LinkedObligationDialog({ children, ...props }: Props) {
  const { __ } = useTranslate();

  return (
    <Dialog trigger={children} title={__("Link obligations")}>
      <DialogContent>
        <Suspense fallback={<Spinner centered />}>
          <LinkedObligationsDialogContent {...props} />
        </Suspense>
      </DialogContent>
      <DialogFooter exitLabel={__("Close")} />
    </Dialog>
  );
}

function LinkedObligationsDialogContent(props: Omit<Props, "children">) {
  const organizationId = useOrganizationId();
  const query = useLazyLoadQuery<LinkedObligationsDialogQuery>(obligationsQuery, {
    organizationId,
  }, { fetchPolicy: "network-only" });
  const { data, loadNext, hasNext, isLoadingNext } = usePaginationFragment(
    obligationsFragment,
    query.organization as LinkedObligationsDialogFragment$key
  );
  const { __ } = useTranslate();
  const [search, setSearch] = useState("");
  const obligations = data.obligations?.edges?.map((edge) => edge.node) ?? [];
  const linkedIds = useMemo(() => {
    return new Set(props.linkedObligations?.map((o) => o.id) ?? []);
  }, [props.linkedObligations]);

  const filteredObligations = useMemo(() => {
    return obligations.filter((obligation) =>
      obligation.area?.toLowerCase().includes(search.toLowerCase()) ||
      obligation.source?.toLowerCase().includes(search.toLowerCase()) ||
      obligation.owner?.fullName?.toLowerCase().includes(search.toLowerCase())
    );
  }, [obligations, search]);

  return (
    <>
      <div className="flex items-center gap-2 sticky top-0 relative py-4 bg-linear-to-b from-50% from-level-2 to-level-2/0 px-6">
        <Input
          icon={IconMagnifyingGlass}
          placeholder={__("Search obligations...")}
          onValueChange={setSearch}
        />
      </div>
      <div className="divide-y divide-border-low">
        {filteredObligations.map((obligation) => (
          <ObligationRow
            key={obligation.id}
            obligation={obligation}
            linkedObligations={linkedIds}
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

type Obligation = NodeOf<LinkedObligationsDialogFragment$data["obligations"]>;

function ObligationRow(props: {
  obligation: Obligation;
  linkedObligations: Set<string>;
  onLink: (obligationId: string) => void;
  onUnlink: (obligationId: string) => void;
  disabled?: boolean;
}) {
  const { __ } = useTranslate();
  const isLinked = props.linkedObligations.has(props.obligation.id);

  const onToggle = () => {
    if (isLinked) {
      props.onUnlink(props.obligation.id);
    } else {
      props.onLink(props.obligation.id);
    }
  };

  return (
    <div className="flex items-center justify-between p-4 hover:bg-level-1">
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-3">
          <div className="flex-1 min-w-0">
            <div className="text-sm font-medium text-txt-primary truncate">
              {props.obligation.area || __("No area specified")}
              {props.obligation.source || __("No source specified")}
            </div>
            <div className="text-xs text-txt-secondary">
              {props.obligation.owner?.fullName || __("Unassigned")}
            </div>
          </div>
          <Badge variant={getObligationStatusVariant(props.obligation.status)}>
            {getObligationStatusLabel(props.obligation.status)}
          </Badge>
        </div>
      </div>
      <Button
        variant={isLinked ? "secondary" : "primary"}
        icon={isLinked ? IconTrashCan : IconPlusLarge}
        onClick={onToggle}
        disabled={props.disabled}
        className="ml-6"
      >
        {isLinked ? __("Unlink") : __("Link")}
      </Button>
    </div>
  );
}
