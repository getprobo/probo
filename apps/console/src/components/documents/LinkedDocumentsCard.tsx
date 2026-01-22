import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  DocumentTypeBadge,
  DocumentVersionBadge,
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

import type { LinkedDocumentsCardFragment$key } from "#/__generated__/core/LinkedDocumentsCardFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { LinkedDocumentDialog } from "./LinkedDocumentsDialog";

const linkedDocumentFragment = graphql`
  fragment LinkedDocumentsCardFragment on Document {
    id
    title
    createdAt
    documentType
    versions(first: 1) {
      edges {
        node {
          id
          status
        }
      }
    }
  }
`;

type Mutation<Params> = (p: {
  variables: {
    input: {
      documentId: string;
    } & Params;
    connections: string[];
  };
}) => void;

type Props<Params> = {
  // Documents linked to the element
  documents: (LinkedDocumentsCardFragment$key & { id: string })[];
  // Extra params to send to the mutation
  params: Params;
  // Disable (action when loading for instance)
  disabled?: boolean;
  // ID of the connection to update
  connectionId: string;
  // Mutation to attach a document (will receive {documentId, ...params})
  onAttach: Mutation<Params>;
  // Mutation to detach a document (will receive {documentId, ...params})
  onDetach: Mutation<Params>;
  variant?: "card" | "table";
  readOnly?: boolean;
};

/**
 * Reusable component that displays a list of linked documents
 */
export function LinkedDocumentsCard<Params>(props: Props<Params>) {
  const { __ } = useTranslate();
  const [limit, setLimit] = useState<number | null>(4);
  const documents = useMemo(() => {
    return limit ? props.documents.slice(0, limit) : props.documents;
  }, [props.documents, limit]);
  const showMoreButton = limit !== null && props.documents.length > limit;
  const variant = props.variant ?? "table";

  const onAttach = (documentId: string) => {
    props.onAttach({
      variables: {
        input: {
          documentId,
          ...props.params,
        },
        connections: [props.connectionId],
      },
    });
  };

  const onDetach = (documentId: string) => {
    props.onDetach({
      variables: {
        input: {
          documentId,
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
          <div className="text-lg font-semibold">{__("Documents")}</div>
          {!props.readOnly && (
            <LinkedDocumentDialog
              connectionId={props.connectionId}
              disabled={props.disabled}
              linkedDocuments={props.documents}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <Button variant="tertiary" icon={IconPlusLarge}>
                {__("Link document")}
              </Button>
            </LinkedDocumentDialog>
          )}
        </div>
      )}
      <Table className={clsx(variant === "card" && "bg-invert")}>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Type")}</Th>
            <Th>{__("State")}</Th>
            {!props.readOnly && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {documents.length === 0 && (
            <Tr>
              <Td
                colSpan={props.readOnly ? 3 : 4}
                className="text-center text-txt-secondary"
              >
                {__("No documents linked")}
              </Td>
            </Tr>
          )}
          {documents.map(document => (
            <DocumentRow
              key={document.id}
              document={document}
              onClick={onDetach}
              readOnly={props.readOnly}
            />
          ))}
          {variant === "table" && !props.readOnly && (
            <LinkedDocumentDialog
              connectionId={props.connectionId}
              disabled={props.disabled}
              linkedDocuments={props.documents}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <TrButton colspan={4} icon={IconPlusLarge}>
                {__("Link document")}
              </TrButton>
            </LinkedDocumentDialog>
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
          {sprintf(__("Show %s more"), props.documents.length - limit)}
        </Button>
      )}
    </Wrapper>
  );
}

function DocumentRow(props: {
  document: LinkedDocumentsCardFragment$key & { id: string };
  onClick: (documentId: string) => void;
  readOnly?: boolean;
}) {
  const document = useFragment(linkedDocumentFragment, props.document);
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  return (
    <Tr to={`/organizations/${organizationId}/documents/${document.id}`}>
      <Td>
        <div className="flex gap-4 items-center">
          <img
            src="/document.png"
            alt=""
            width={28}
            height={36}
            className="border-4 border-highlight rounded box-content"
          />
          {document.title}
        </div>
      </Td>
      <Td>
        <DocumentTypeBadge type={document.documentType} />
      </Td>
      <Td>
        <DocumentVersionBadge state={document.versions.edges[0].node.status} />
      </Td>
      {!props.readOnly && (
        <Td noLink width={50} className="text-end">
          <Button
            variant="secondary"
            onClick={() => props.onClick(document.id)}
            icon={IconTrashCan}
          >
            {__("Unlink")}
          </Button>
        </Td>
      )}
    </Tr>
  );
}
