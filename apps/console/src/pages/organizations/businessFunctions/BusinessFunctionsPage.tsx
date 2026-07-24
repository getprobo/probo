// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Avatar,
  Button,
  Card,
  Checkbox,
  IconPageTextLine,
  IconPlusLarge,
  IconUpload,
  Option,
  PageHeader,
  Select,
  Table,
  Tbody,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { Suspense, useState, useTransition } from "react";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { Link, useNavigate } from "react-router";

import type { BusinessFunctionsPageFragment$key } from "#/__generated__/core/BusinessFunctionsPageFragment.graphql";
import type { BusinessFunctionsPageListQuery } from "#/__generated__/core/BusinessFunctionsPageListQuery.graphql";
import type {
  BusinessFunctionClassification,
  BusinessFunctionsPageRefetchQuery,
} from "#/__generated__/core/BusinessFunctionsPageRefetchQuery.graphql";
import { usePeople } from "#/hooks/graph/PeopleGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { BusinessFunctionListItem } from "./_components/BusinessFunctionListItem";
import {
  BusinessFunctionsConnectionKey,
  emptyBusinessFunctionFilter,
} from "./_lib/businessFunctionHelpers";
import { CreateBusinessFunctionDialog } from "./dialogs/CreateBusinessFunctionDialog";
import { PublishBusinessFunctionListDialog } from "./dialogs/PublishBusinessFunctionListDialog";

export const businessFunctionsPageQuery = graphql`
  query BusinessFunctionsPageListQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        canCreateBusinessFunction: permission(action: "core:business-function:create")
        canPublishBusinessFunctions: permission(action: "core:business-function:publish")
        businessFunctionsDocument {
          id
          defaultApprovers {
            id
          }
        }
        ...BusinessFunctionsPageFragment
      }
    }
  }
`;

const businessFunctionsPageFragment = graphql`
  fragment BusinessFunctionsPageFragment on Organization
  @refetchable(queryName: "BusinessFunctionsPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 500 }
    after: { type: "CursorKey" }
    classification: { type: "BusinessFunctionClassification", defaultValue: null }
    ownerId: { type: "ID", defaultValue: null }
    cifOnly: { type: "Boolean", defaultValue: null }
  ) {
    id
    businessFunctions(
      first: $first
      after: $after
      filter: {
        classification: $classification
        ownerId: $ownerId
        cifOnly: $cifOnly
      }
    )
      @connection(
        key: "BusinessFunctionsPage_businessFunctions"
        filters: ["filter"]
      ) {
      edges {
        node {
          id
          canDelete: permission(action: "core:business-function:delete")
          ...BusinessFunctionListItem_businessFunction
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

interface BusinessFunctionsPageProps {
  queryRef: PreloadedQuery<BusinessFunctionsPageListQuery>;
}

export default function BusinessFunctionsPage({ queryRef }: BusinessFunctionsPageProps) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  usePageTitle(__("Business functions"));

  const navigate = useNavigate();
  const organization = usePreloadedQuery<BusinessFunctionsPageListQuery>(
    businessFunctionsPageQuery,
    queryRef,
  );
  const defaultApproverIds = (
    organization.node.businessFunctionsDocument?.defaultApprovers ?? []
  ).map(approver => approver.id);

  const [isPending, startTransition] = useTransition();
  const [classificationFilter, setClassificationFilter]
    = useState<BusinessFunctionClassification | null>(null);
  const [ownerFilter, setOwnerFilter] = useState<string | null>(null);
  const [cifOnlyFilter, setCifOnlyFilter] = useState<boolean>(false);

  const { data, loadNext, hasNext, isLoadingNext, refetch }
    = usePaginationFragment<
      BusinessFunctionsPageRefetchQuery,
      BusinessFunctionsPageFragment$key
    >(businessFunctionsPageFragment, organization.node);

  const refetchFilters = (overrides: Record<string, unknown> = {}) => {
    startTransition(() => {
      refetch(
        {
          classification: classificationFilter,
          ownerId: ownerFilter,
          cifOnly: cifOnlyFilter ? true : null,
          ...overrides,
        },
        { fetchPolicy: "network-only" },
      );
    });
  };

  const handleClassificationFilterChange = (value: string) => {
    const newClassification = value === "ALL"
      ? null
      : (value as BusinessFunctionClassification);
    setClassificationFilter(newClassification);
    refetchFilters({ classification: newClassification });
  };

  const handleOwnerFilterChange = (value: string) => {
    const newOwner = value === "ALL" ? null : value;
    setOwnerFilter(newOwner);
    refetchFilters({ ownerId: newOwner });
  };

  const handleCifOnlyFilterChange = (checked: boolean) => {
    setCifOnlyFilter(checked);
    refetchFilters({ cifOnly: checked ? true : null });
  };

  const allFiltersNullConnectionId = ConnectionHandler.getConnectionID(
    organizationId,
    BusinessFunctionsConnectionKey,
    { filter: emptyBusinessFunctionFilter },
  );
  // Only prepend into the unfiltered connection so filtered views stay accurate.
  const createConnectionIds = [allFiltersNullConnectionId];
  const businessFunctions = data?.businessFunctions?.edges?.map(edge => edge.node) ?? [];

  const hasAnyAction = businessFunctions.some(({ canDelete }) => canDelete);

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Business functions")}
        description={__("Manage your organization's business functions register.")}
      >
        <div className="flex gap-2">
          {organization.node.businessFunctionsDocument?.id && (
            <Button variant="secondary" asChild>
              <Link
                to={`/organizations/${organizationId}/documents/${organization.node.businessFunctionsDocument.id}`}
              >
                <IconPageTextLine size={16} />
                {__("Document")}
              </Link>
            </Button>
          )}
          {organization.node.canPublishBusinessFunctions && (
            <PublishBusinessFunctionListDialog
              organizationId={organizationId}
              defaultApproverIds={defaultApproverIds}
              onPublished={(documentId) => {
                void navigate(
                  `/organizations/${organizationId}/documents/${documentId}`,
                );
              }}
            >
              <Button variant="secondary" icon={IconUpload}>
                {__("Publish")}
              </Button>
            </PublishBusinessFunctionListDialog>
          )}
          {organization.node.canCreateBusinessFunction && (
            <CreateBusinessFunctionDialog
              organizationId={organizationId}
              connectionIds={createConnectionIds}
            >
              <Button icon={IconPlusLarge}>{__("Add business function")}</Button>
            </CreateBusinessFunctionDialog>
          )}
        </div>
      </PageHeader>

      <div className="flex flex-wrap items-center gap-4">
        <Select
          value={classificationFilter ?? "ALL"}
          onValueChange={handleClassificationFilterChange}
        >
          <Option value="ALL">{__("All classifications")}</Option>
          <Option value="CRITICAL">{__("Critical")}</Option>
          <Option value="IMPORTANT">{__("Important")}</Option>
          <Option value="SECONDARY">{__("Secondary")}</Option>
          <Option value="STANDARD">{__("Standard")}</Option>
        </Select>
        <Suspense fallback={<Select loading placeholder={__("Loading...")} />}>
          <OwnerFilterSelect
            organizationId={organizationId}
            value={ownerFilter}
            onChange={handleOwnerFilterChange}
          />
        </Suspense>
        <div className="flex items-center gap-2">
          <Checkbox
            checked={cifOnlyFilter}
            onChange={handleCifOnlyFilterChange}
          />
          <button
            type="button"
            className="text-sm font-medium text-txt-primary"
            onClick={() => handleCifOnlyFilterChange(!cifOnlyFilter)}
          >
            {__("CIF only (Critical + Important)")}
          </button>
        </div>
      </div>

      <div className={isPending ? "opacity-50 pointer-events-none transition-opacity" : ""}>
        {businessFunctions.length > 0
          ? (
              <Card>
                <Table>
                  <Thead>
                    <Tr>
                      <Th>{__("Reference ID")}</Th>
                      <Th>{__("Name")}</Th>
                      <Th>{__("Classification")}</Th>
                      <Th>{__("MTD (min)")}</Th>
                      <Th>{__("RTO (min)")}</Th>
                      <Th>{__("RPO (min)")}</Th>
                      <Th>{__("Owner")}</Th>
                      {hasAnyAction && <Th>{__("Actions")}</Th>}
                    </Tr>
                  </Thead>
                  <Tbody>
                    {businessFunctions.map(businessFunction => (
                      <BusinessFunctionListItem
                        key={businessFunction.id}
                        businessFunctionKey={businessFunction}
                        hasAnyAction={hasAnyAction}
                      />
                    ))}
                  </Tbody>
                </Table>

                {hasNext && (
                  <div className="p-4 border-t">
                    <Button
                      variant="secondary"
                      onClick={() => loadNext(10)}
                      disabled={isLoadingNext}
                    >
                      {isLoadingNext ? __("Loading...") : __("Load more")}
                    </Button>
                  </div>
                )}
              </Card>
            )
          : (
              <Card padded>
                <div className="text-center py-12">
                  <h3 className="text-lg font-semibold mb-2">
                    {__("No business functions yet")}
                  </h3>
                  <p className="text-txt-tertiary mb-4">
                    {__("Create your first business function to get started.")}
                  </p>
                </div>
              </Card>
            )}
      </div>
    </div>
  );
}

type OwnerFilterSelectProps = {
  organizationId: string;
  value: string | null;
  onChange: (value: string) => void;
};

function OwnerFilterSelect({
  organizationId,
  value,
  onChange,
}: OwnerFilterSelectProps) {
  const { __ } = useTranslate();
  const people = usePeople(organizationId, { contractEnded: false });

  return (
    <Select value={value ?? "ALL"} onValueChange={onChange}>
      <Option value="ALL">{__("All owners")}</Option>
      {people.map(person => (
        <Option key={person.id} value={person.id}>
          <Avatar name={person.fullName} />
          {person.fullName}
        </Option>
      ))}
    </Select>
  );
}
