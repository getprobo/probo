// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { LinkedDocumentsCardFragment$key } from "#/__generated__/core/LinkedDocumentsCardFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { LinkedDocumentDialog } from "./LinkedDocumentsDialog";

const linkedDocumentFragment = graphql`
  fragment LinkedDocumentsCardFragment on Document {
    id
    versions(first: 1) {
      edges {
        node {
          id
          title
          documentType
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
  const { t } = useTranslation();
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
          <div className="text-lg font-semibold">{t("linkedDocumentsCard.title")}</div>
          {!props.readOnly && (
            <LinkedDocumentDialog
              connectionId={props.connectionId}
              disabled={props.disabled}
              linkedDocuments={props.documents}
              onLink={onAttach}
              onUnlink={onDetach}
            >
              <Button variant="tertiary" icon={IconPlusLarge}>
                {t("linkedDocumentsCard.actions.link")}
              </Button>
            </LinkedDocumentDialog>
          )}
        </div>
      )}
      <Table className={clsx(variant === "card" && "bg-invert")}>
        <Thead>
          <Tr>
            <Th>{t("linkedDocumentsCard.columns.name")}</Th>
            <Th>{t("linkedDocumentsCard.columns.type")}</Th>
            <Th>{t("linkedDocumentsCard.columns.state")}</Th>
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
                {t("linkedDocumentsCard.empty")}
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
                {t("linkedDocumentsCard.actions.link")}
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
          {t("linkedDocumentsCard.actions.showMore", {
            count: props.documents.length - limit,
          })}
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
  const { t } = useTranslation();

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
          {document.versions.edges[0].node.title}
        </div>
      </Td>
      <Td>
        <DocumentTypeBadge type={document.versions.edges[0].node.documentType} />
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
            {t("linkedDocumentsCard.actions.unlink")}
          </Button>
        </Td>
      )}
    </Tr>
  );
}
