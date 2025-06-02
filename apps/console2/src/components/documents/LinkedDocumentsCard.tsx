import { graphql } from "relay-runtime";
import {
  Card,
  IconPlusLarge,
  Button,
  Tr,
  Td,
  Table,
  Thead,
  Tbody,
  Th,
  IconChevronDown,
  IconTrashCan,
  DocumentVersionBadge,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { LinkedDocumentsCardFragment$key } from "./__generated__/LinkedDocumentsCardFragment.graphql";
import { useFragment } from "react-relay";
import { useMemo, useState } from "react";
import { sprintf } from "@probo/helpers";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { DocumentLinkDialog } from "./DocumentLinkDialog";

const linkedDocumentFragment = graphql`
  fragment LinkedDocumentsCardFragment on Document {
    id
    title
    createdAt
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

  return (
    <Card padded className="space-y-[10px]">
      <div className="flex justify-between">
        <div className="text-lg font-semibold">{__("Documents")}</div>
        <DocumentLinkDialog
          connectionId={props.connectionId}
          disabled={props.disabled}
          linkedDocuments={props.documents}
          onLink={onAttach}
          onUnlink={onDetach}
        >
          <Button variant="tertiary" icon={IconPlusLarge}>
            {__("Link document")}
          </Button>
        </DocumentLinkDialog>
      </div>
      {documents.length > 0 ? (
        <Table className="bg-invert">
          <Thead>
            <Tr>
              <Th>{__("Name")}</Th>
              <Th>{__("State")}</Th>
              <Th></Th>
            </Tr>
          </Thead>
          <Tbody>
            {documents.map((document) => (
              <DocumentRow
                key={document.id}
                document={document}
                onClick={onDetach}
              />
            ))}
          </Tbody>
        </Table>
      ) : (
        <div className="text-center text-sm text-txt-secondary">
          {__("No documents linked")}
        </div>
      )}
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
    </Card>
  );
}

function DocumentRow(props: {
  document: LinkedDocumentsCardFragment$key & { id: string };
  onClick: (documentId: string) => void;
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
        <DocumentVersionBadge state={document.versions.edges[0].node.status} />
      </Td>
      <Td noLink width={50} className="text-end">
        <Button
          variant="secondary"
          onClick={() => props.onClick(document.id)}
          icon={IconTrashCan}
        >
          {__("Unlink")}
        </Button>
      </Td>
    </Tr>
  );
}
