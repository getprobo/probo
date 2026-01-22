import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DocumentTypeBadge,
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
  LinkedDocumentsDialogFragment$data,
  LinkedDocumentsDialogFragment$key,
} from "#/__generated__/core/LinkedDocumentsDialogFragment.graphql";
import type { LinkedDocumentsDialogQuery } from "#/__generated__/core/LinkedDocumentsDialogQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

const documentsQuery = graphql`
  query LinkedDocumentsDialogQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        ...LinkedDocumentsDialogFragment
      }
    }
  }
`;

const documentsFragment = graphql`
  fragment LinkedDocumentsDialogFragment on Organization
  @refetchable(queryName: "LinkedDocumentsDialogQuery_fragment")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    order: { type: "DocumentOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    documents(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "LinkedDocumentsDialogQuery_documents") {
      edges {
        node {
          id
          title
          documentType
        }
      }
    }
  }
`;

type Props = {
  children: ReactNode;
  connectionId: string;
  disabled?: boolean;
  linkedDocuments?: { id: string }[];
  onLink: (documentId: string) => void;
  onUnlink: (documentId: string) => void;
};

export function LinkedDocumentDialog({ children, ...props }: Props) {
  const { __ } = useTranslate();

  return (
    <Dialog trigger={children} title={__("Link documents")}>
      <DialogContent>
        <Suspense fallback={<Spinner centered />}>
          <LinkedDocumentsDialogContent {...props} />
        </Suspense>
      </DialogContent>
      <DialogFooter exitLabel={__("Close")} />
    </Dialog>
  );
}

function LinkedDocumentsDialogContent(props: Omit<Props, "children">) {
  const organizationId = useOrganizationId();
  const query = useLazyLoadQuery<LinkedDocumentsDialogQuery>(
    documentsQuery,
    {
      organizationId,
    },
    { fetchPolicy: "network-only" },
  );
  const { data, loadNext, hasNext, isLoadingNext } = usePaginationFragment(
    documentsFragment,
    query.organization as LinkedDocumentsDialogFragment$key,
  );
  const { __ } = useTranslate();
  const [search, setSearch] = useState("");
  const documents = useMemo(
    () => data.documents?.edges?.map(edge => edge.node) ?? [],
    [data.documents],
  );
  const linkedIds = useMemo(() => {
    return new Set(props.linkedDocuments?.map(m => m.id) ?? []);
  }, [props.linkedDocuments]);

  const filteredDocuments = useMemo(() => {
    return documents.filter(document =>
      document.title.toLowerCase().includes(search.toLowerCase()),
    );
  }, [documents, search]);

  return (
    <>
      <div className="flex items-center gap-2 sticky top-0 relative py-4 bg-linear-to-b from-50% from-level-2 to-level-2/0 px-6">
        <Input
          icon={IconMagnifyingGlass}
          placeholder={__("Search documents...")}
          onValueChange={setSearch}
        />
      </div>
      <div className="divide-y divide-border-low">
        {filteredDocuments.map(document => (
          <DocumentRow
            key={document.id}
            document={document}
            linkedDocuments={linkedIds}
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

type Document = NodeOf<LinkedDocumentsDialogFragment$data["documents"]>;

type RowProps = {
  document: Document;
  linkedDocuments: Set<string>;
  disabled?: boolean;
  onLink: (documentId: string) => void;
  onUnlink: (documentId: string) => void;
};

function DocumentRow(props: RowProps) {
  const { __ } = useTranslate();

  const isLinked = props.linkedDocuments.has(props.document.id);
  const onClick = isLinked ? props.onUnlink : props.onLink;
  const IconComponent = isLinked ? IconTrashCan : IconPlusLarge;

  return (
    <button
      className="py-4 flex items-center gap-4 hover:bg-subtle cursor-pointer px-6 w-full h-[100px]"
      onClick={() => onClick(props.document.id)}
    >
      {props.document.title}
      <DocumentTypeBadge type={props.document.documentType} />
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
