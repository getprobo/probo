import { useTranslate } from "@probo/i18n";
import {
  PageHeader,
  Tbody,
  Thead,
  Tr,
  Th,
  Td,
  Badge,
  Card,
} from "@probo/ui";
import { SortableTable } from "/components/SortableTable";
import {
  useFragment,
  usePaginationFragment,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import { graphql } from "relay-runtime";
import type { EmployeeDocumentsPageListQuery } from "./__generated__/EmployeeDocumentsPageListQuery.graphql";
import type { EmployeeDocumentsPageListFragment$key } from "./__generated__/EmployeeDocumentsPageListFragment.graphql";
import { usePageTitle } from "@probo/hooks";
import { getDocumentClassificationLabel, getDocumentTypeLabel, formatDate } from "@probo/helpers";
import type { EmployeeDocumentsPageRowFragment$key } from "./__generated__/EmployeeDocumentsPageRowFragment.graphql";
import { useEffect } from "react";

export const employeeDocumentsQuery = graphql`
  query EmployeeDocumentsPageListQuery($organizationId: ID!) {
    viewer {
      id
      ...EmployeeDocumentsPageListFragment @arguments(organizationId: $organizationId)
    }
  }
`;

const employeeDocumentsFragment = graphql`
  fragment EmployeeDocumentsPageListFragment on Viewer
  @refetchable(queryName: "EmployeeDocumentsListQuery")
  @argumentDefinitions(
    organizationId: { type: "ID!" }
    first: { type: "Int", defaultValue: 50 }
    order: {
      type: "DocumentOrder"
      defaultValue: { field: CREATED_AT, direction: DESC }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    signableDocuments(
      organizationId: $organizationId
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "EmployeeDocumentsListQuery_signableDocuments") {
      __id
      edges {
        node {
          id
          ...EmployeeDocumentsPageRowFragment
        }
      }
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<EmployeeDocumentsPageListQuery>;
  organizationId: string;
};

export default function EmployeeDocumentsPage(props: Props) {
  const { __ } = useTranslate();
  const organizationId = props.organizationId;

  const data = usePreloadedQuery(
    employeeDocumentsQuery,
    props.queryRef
  );

  const pagination = usePaginationFragment(
    employeeDocumentsFragment,
    data.viewer as EmployeeDocumentsPageListFragment$key
  );

  useEffect(() => {
    pagination.refetch({ organizationId }, { fetchPolicy: 'network-only' });
  }, [organizationId]);

  const documents = pagination.data.signableDocuments?.edges
    ?.map((edge) => edge?.node)
    .filter(Boolean) || [];

  usePageTitle(__("Documents"));

  return (
    <div className="space-y-6">
      <PageHeader title={__("Documents")} />
      {documents.length > 0 ? (
        <SortableTable {...pagination}>
          <Thead>
          <Tr>
            <Th className="min-w-0 pr-12">{__("Name")}</Th>
            <Th className="w-48">{__("Type")}</Th>
            <Th className="w-36">{__("Classification")}</Th>
            <Th className="w-40">{__("Last update")}</Th>
            <Th className="w-32">{__("Signed")}</Th>
          </Tr>
          </Thead>
          <Tbody>
                {documents.map((document) => (
                  <DocumentRow key={document.id} document={document} organizationId={organizationId} />
                ))}
          </Tbody>
        </SortableTable>
      ) : (
        <Card padded>
          <div className="text-center py-12">
            <h3 className="text-lg font-semibold mb-2">
              {__("No documents yet")}
            </h3>
            <p className="text-txt-tertiary mb-4">
              {__("No documents have been requested for your signature.")}
            </p>
          </div>
        </Card>
      )}
    </div>
  );
}

const rowFragment = graphql`
  fragment EmployeeDocumentsPageRowFragment on SignableDocument {
    id
    title
    documentType
    classification
    signed
    updatedAt
  }
`;

function DocumentRow({
  document: documentKey,
  organizationId,
}: {
  document: EmployeeDocumentsPageRowFragment$key;
  organizationId: string;
}) {
  const document = useFragment<EmployeeDocumentsPageRowFragment$key>(
    rowFragment,
    documentKey
  );
  const { __ } = useTranslate();

  return (
    <Tr
      to={`/organizations/${organizationId}/employee/${document.id}`}
    >
      <Td className="min-w-0 pr-12">{document.title}</Td>
      <Td className="w-48">{getDocumentTypeLabel(__, document.documentType)}</Td>
      <Td className="w-36">
        <Badge variant="neutral">
          {getDocumentClassificationLabel(__, document.classification)}
        </Badge>
      </Td>
      <Td className="w-40">{formatDate(document.updatedAt)}</Td>
      <Td className="w-32">
        <Badge variant={document.signed ? "success" : "danger"}>
          {document.signed ? __("Yes") : __("No")}
        </Badge>
      </Td>
    </Tr>
  );
}
