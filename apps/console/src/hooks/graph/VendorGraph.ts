import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "../useMutationWithToasts.ts";
import type { VendorGraphCreateMutation } from "./__generated__/VendorGraphCreateMutation.graphql.ts";
import type { VendorGraphDeleteMutation } from "./__generated__/VendorGraphDeleteMutation.graphql.ts";
import type { VendorGraphSelectQuery } from "./__generated__/VendorGraphSelectQuery.graphql.ts";
import { useMutation, useLazyLoadQuery } from "react-relay";
import { useConfirm } from "@probo/ui";
import { promisifyMutation, sprintf } from "@probo/helpers";
import { useMemo } from "react";

const createVendorMutation = graphql`
  mutation VendorGraphCreateMutation(
    $input: CreateVendorInput!
    $connections: [ID!]!
  ) {
    createVendor(input: $input) {
      vendorEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          description
          websiteUrl
          createdAt
          updatedAt
        }
      }
    }
  }
`;

export function useCreateVendorMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<VendorGraphCreateMutation>(
    createVendorMutation,
    {
      successMessage: __("Vendor created successfully."),
      errorMessage: __("Failed to create vendor. Please try again."),
    }
  );
}

const deleteVendorMutation = graphql`
  mutation VendorGraphDeleteMutation(
    $input: DeleteVendorInput!
    $connections: [ID!]!
  ) {
    deleteVendor(input: $input) {
      deletedVendorId @deleteEdge(connections: $connections)
    }
  }
`;

export const useDeleteVendor = (
  vendor: { id?: string; name?: string },
  connectionId: string
) => {
  const [mutate] = useMutation<VendorGraphDeleteMutation>(deleteVendorMutation);
  const confirm = useConfirm();
  const { __ } = useTranslate();

  return () => {
    if (!vendor.id || !vendor.name) {
      return alert(__("Failed to delete vendor: missing id or name"));
    }
    confirm(
      () =>
        promisifyMutation(mutate)({
          variables: {
            input: {
              vendorId: vendor.id!,
            },
            connections: [connectionId],
          },
        }),
      {
        message: sprintf(
          __(
            'This will permanently delete vendor "%s". This action cannot be undone.'
          ),
          vendor.name
        ),
      }
    );
  };
};

export const vendorConnectionKey = "VendorsPage_vendors";

export const vendorsQuery = graphql`
  query VendorGraphListQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        id
        ...VendorGraphPaginatedFragment
      }
    }
  }
`;

export const paginatedVendorsFragment = graphql`
  fragment VendorGraphPaginatedFragment on Organization
  @refetchable(queryName: "VendorsListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "VendorOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    vendors(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "VendorsListQuery_vendors") {
      __id
      edges {
        node {
          id
          name
          websiteUrl
          updatedAt
          riskAssessments(
            first: 1
            orderBy: { direction: DESC, field: ASSESSED_AT }
          ) {
            edges {
              node {
                id
                assessedAt
                expiresAt
                dataSensitivity
                businessImpact
              }
            }
          }
        }
      }
    }
  }
`;

export const vendorNodeQuery = graphql`
  query VendorGraphNodeQuery($vendorId: ID!, $organizationId: ID!) {
    node(id: $vendorId) {
      ... on Vendor {
        id
        name
        websiteUrl
        ...useVendorFormFragment
        ...VendorComplianceTabFragment
        ...VendorContactsTabFragment
        ...VendorServicesTabFragment
        ...VendorRiskAssessmentTabFragment
        ...VendorOverviewTabBusinessAssociateAgreementFragment
        ...VendorOverviewTabDataPrivacyAgreementFragment
      }
    }
    viewer {
      user {
        people(organizationId: $organizationId) {
          id
        }
      }
    }
  }
`;

const vendorsSelectQuery = graphql`
  query VendorGraphSelectQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        vendors(first: 100, orderBy: { direction: ASC, field: NAME }) {
          edges {
            node {
              id
              name
              websiteUrl
            }
          }
        }
      }
    }
  }
`;

export function useVendors(organizationId: string) {
  const data = useLazyLoadQuery<VendorGraphSelectQuery>(
    vendorsSelectQuery,
    {
      organizationId: organizationId,
    },
    { fetchPolicy: "network-only" }
  );
  return useMemo(() => {
    return data.organization?.vendors?.edges.map((edge) => edge.node) ?? [];
  }, [data]);
}
