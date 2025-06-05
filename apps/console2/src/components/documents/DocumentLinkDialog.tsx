import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DocumentTypeBadge,
  IconMagnifyingGlass,
  IconPlusLarge,
  IconTrashCan,
  Input,
  Spinner,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { Suspense, useMemo, useState, type ReactNode } from "react";
import { graphql } from "relay-runtime";
import { useLazyLoadQuery } from "react-relay";
import type {
  DocumentLinkDialogQuery,
  DocumentLinkDialogQuery$data,
} from "./__generated__/DocumentLinkDialogQuery.graphql";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { NodeOf } from "/types";

const documentsQuery = graphql`
  query DocumentLinkDialogQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        documents(first: 100) @connection(key: "Organization__documents") {
          edges {
            node {
              id
              title
              documentType
            }
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
  linkedDocuments?: { id: string }[];
  onLink: (documentId: string) => void;
  onUnlink: (documentId: string) => void;
};

export function DocumentLinkDialog({ children, ...props }: Props) {
  const { __ } = useTranslate();

  return (
    <Dialog trigger={children} title={__("Link documents")}>
      <DialogContent>
        <Suspense fallback={<Spinner centered />}>
          <DocumentLinkDialogContent {...props} />
        </Suspense>
      </DialogContent>
      <DialogFooter exitLabel={__("Close")} />
    </Dialog>
  );
}

function DocumentLinkDialogContent(props: Omit<Props, "children">) {
  const organizationId = useOrganizationId();
  const data = useLazyLoadQuery<DocumentLinkDialogQuery>(documentsQuery, {
    organizationId,
  });
  const { __ } = useTranslate();
  const [search, setSearch] = useState("");
  const documents =
    data.organization?.documents?.edges?.map((edge) => edge.node) ?? [];
  const linkedIds = useMemo(() => {
    return new Set(props.linkedDocuments?.map((m) => m.id) ?? []);
  }, [props.linkedDocuments]);

  const filteredDocuments = useMemo(() => {
    return documents.filter((document) =>
      document.title.toLowerCase().includes(search.toLowerCase())
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
        {filteredDocuments.map((document) => (
          <DocumentRow
            key={document.id}
            document={document}
            linkedDocuments={linkedIds}
            onLink={props.onLink}
            onUnlink={props.onUnlink}
            disabled={props.disabled}
          />
        ))}
      </div>
    </>
  );
}

type Document = NodeOf<
  DocumentLinkDialogQuery$data["organization"]["documents"]
>;

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
      className="py-4 flex items-center gap-4 hover:bg-subtle cursor-pointer px-6 w-full"
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
          <IconComponent size={16} /> {isLinked ? __("Unlink") : __("Link")}
        </span>
      </Button>
    </button>
  );
}
