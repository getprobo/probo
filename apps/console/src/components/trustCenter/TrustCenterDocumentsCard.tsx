import { graphql } from "relay-runtime";
import {
  Card,
  Button,
  Tr,
  Td,
  Table,
  Thead,
  Tbody,
  Th,
  IconChevronDown,
  IconCheckmark1,
  IconCrossLargeX,
  DocumentVersionBadge,
  DocumentTypeBadge,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { TrustCenterDocumentsCardFragment$key } from "./__generated__/TrustCenterDocumentsCardFragment.graphql";
import { useFragment } from "react-relay";
import { useMemo, useState } from "react";
import { sprintf } from "@probo/helpers";
import { useOrganizationId } from "/hooks/useOrganizationId";
import clsx from "clsx";

const trustCenterDocumentFragment = graphql`
  fragment TrustCenterDocumentsCardFragment on Document {
    id
    title
    createdAt
    documentType
    showOnTrustCenter
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
      id: string;
      showOnTrustCenter: boolean;
    } & Params;
  };
}) => void;

type Props<Params> = {
  documents: (TrustCenterDocumentsCardFragment$key & { id: string })[];
  params: Params;
  disabled?: boolean;
  onToggleVisibility: Mutation<Params>;
  variant?: "card" | "table";
};

export function TrustCenterDocumentsCard<Params>(props: Props<Params>) {
  const { __ } = useTranslate();
  const [limit, setLimit] = useState<number | null>(4);
  const documents = useMemo(() => {
    return limit ? props.documents.slice(0, limit) : props.documents;
  }, [props.documents, limit]);
  const showMoreButton = limit !== null && props.documents.length > limit;
  const variant = props.variant ?? "table";

  const onToggleVisibility = (documentId: string, showOnTrustCenter: boolean) => {
    props.onToggleVisibility({
      variables: {
        input: {
          id: documentId,
          showOnTrustCenter,
          ...props.params,
        },
      },
    });
  };

  const Wrapper = variant === "card" ? Card : "div";

  return (
    <Wrapper padded className="space-y-[10px]">
      <Table className={clsx(variant === "card" && "bg-invert")}>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Type")}</Th>
            <Th>{__("State")}</Th>
            <Th>{__("Visibility")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {documents.length === 0 && (
            <Tr>
              <Td colSpan={5} className="text-center text-txt-secondary">
                {__("No documents available")}
              </Td>
            </Tr>
          )}
          {documents.map((document) => (
            <DocumentRow
              key={document.id}
              document={document}
              onToggleVisibility={onToggleVisibility}
              disabled={props.disabled}
            />
          ))}

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
  document: TrustCenterDocumentsCardFragment$key & { id: string };
  onToggleVisibility: (documentId: string, showOnTrustCenter: boolean) => void;
  disabled?: boolean;
}) {
  const document = useFragment(trustCenterDocumentFragment, props.document);
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  return (
    <Tr to={`/organizations/${organizationId}/documents/${document.id}`}>
      <Td>
        <div className="flex gap-4 items-center">
          {document.title}
        </div>
      </Td>
      <Td>
        <DocumentTypeBadge type={document.documentType} />
      </Td>
      <Td>
        <DocumentVersionBadge state={document.versions?.edges?.[0]?.node?.status} />
      </Td>
      <Td>
        <div className="flex items-center gap-2">
          {document.showOnTrustCenter ? (
            <>
              <IconCheckmark1 className="w-4 h-4 text-txt-primary" />
              <span className="text-txt-primary">{__("Visible")}</span>
            </>
          ) : (
            <>
              <IconCrossLargeX className="w-4 h-4 text-txt-tertiary" />
              <span className="text-txt-tertiary">{__("Hidden")}</span>
            </>
          )}
        </div>
      </Td>
      <Td noLink width={100} className="text-end">
        <Button
          variant="secondary"
          onClick={() => props.onToggleVisibility(document.id, !document.showOnTrustCenter)}
          icon={document.showOnTrustCenter ? IconCrossLargeX : IconCheckmark1}
          disabled={props.disabled}
        >
          {document.showOnTrustCenter ? __("Hide") : __("Show")}
        </Button>
      </Td>
    </Tr>
  );
}
