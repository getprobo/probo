import { graphql } from "relay-runtime";
import {
  Button,
  Tr,
  Td,
  Table,
  Thead,
  Tbody,
  Th,
  IconChevronDown,
  DocumentVersionBadge,
  DocumentTypeBadge,
  Field,
  Option,
  Badge,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { TrustCenterDocumentsCardFragment$key } from "./__generated__/TrustCenterDocumentsCardFragment.graphql";
import { useFragment } from "react-relay";
import { useMemo, useState, useCallback, useEffect } from "react";
import { sprintf, getTrustCenterVisibilityOptions } from "@probo/helpers";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { isAuthorized } from "/permissions";

const trustCenterDocumentFragment = graphql`
  fragment TrustCenterDocumentsCardFragment on Document {
    id
    title
    createdAt
    documentType
    trustCenterVisibility
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
      trustCenterVisibility: "NONE" | "PRIVATE" | "PUBLIC";
    } & Params;
  };
}) => void;

type Props<Params> = {
  documents: TrustCenterDocumentsCardFragment$key[];
  params: Params;
  disabled?: boolean;
  onChangeVisibility: Mutation<Params>;
};

export function TrustCenterDocumentsCard<Params>(props: Props<Params>) {
  const { __ } = useTranslate();
  const [limit, setLimit] = useState<number | null>(100);
  const documents = useMemo(() => {
    return limit ? props.documents.slice(0, limit) : props.documents;
  }, [props.documents, limit]);
  const showMoreButton = limit !== null && props.documents.length > limit;

  const onChangeVisibility = (documentId: string, trustCenterVisibility: "NONE" | "PRIVATE" | "PUBLIC") => {
    props.onChangeVisibility({
      variables: {
        input: {
          id: documentId,
          trustCenterVisibility,
          ...props.params,
        },
      },
    });
  };

  return (
    <div className="space-y-[10px]">
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Type")}</Th>
            <Th>{__("State")}</Th>
            <Th>{__("Visibility")}</Th>
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
          {documents.map((document, index) => (
            <DocumentRow
              key={index}
              document={document}
              onChangeVisibility={onChangeVisibility}
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
    </div>
  );
}

function DocumentRow(props: {
  document: TrustCenterDocumentsCardFragment$key;
  onChangeVisibility: (documentId: string, trustCenterVisibility: "NONE" | "PRIVATE" | "PUBLIC") => void;
  disabled?: boolean;
}) {
  const document = useFragment(trustCenterDocumentFragment, props.document);
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const [optimisticValue, setOptimisticValue] = useState<string | null>(null);

  const canUpdate = organizationId ? isAuthorized(organizationId, "TrustCenter", "updateTrustCenter") : false;

  const handleValueChange = useCallback((value: string | {}) => {
    const stringValue = typeof value === 'string' ? value : '';
    const typedValue = stringValue as "NONE" | "PRIVATE" | "PUBLIC";
    setOptimisticValue(typedValue);
    props.onChangeVisibility(document.id, typedValue);
  }, [document.id, props.onChangeVisibility]);

  useEffect(() => {
    if (optimisticValue && document.trustCenterVisibility === optimisticValue) {
      setOptimisticValue(null);
    }
  }, [document.trustCenterVisibility, optimisticValue]);

  const currentValue = optimisticValue || document.trustCenterVisibility;

  const visibilityOptions = getTrustCenterVisibilityOptions(__);

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
      <Td noLink width={130} className="pr-0">
        <Field
          type="select"
          value={currentValue}
          onValueChange={handleValueChange}
          disabled={props.disabled || !canUpdate}
          className="w-[105px]"
        >
          {visibilityOptions.map((option) => (
            <Option key={option.value} value={option.value}>
              <div className="flex items-center justify-between w-full">
                <Badge variant={option.variant}>
                  {option.label}
                </Badge>
              </div>
            </Option>
          ))}
        </Field>
      </Td>
    </Tr>
  );
}
