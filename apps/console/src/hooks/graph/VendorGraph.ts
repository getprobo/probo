import { promisifyMutation, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useConfirm } from "@probo/ui";
import { useMemo } from "react";
import { useLazyLoadQuery, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { VendorGraphCreateMutation } from "#/__generated__/core/VendorGraphCreateMutation.graphql";
import type { VendorGraphDeleteMutation } from "#/__generated__/core/VendorGraphDeleteMutation.graphql";
import type { VendorGraphSelectQuery } from "#/__generated__/core/VendorGraphSelectQuery.graphql";

import { useMutationWithToasts } from "../useMutationWithToasts";

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
          canUpdate: permission(action: "core:vendor:update")
          canDelete: permission(action: "core:vendor:delete")
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
      errorMessage: __("Failed to create vendor"),
    },
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
  connectionId: string,
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
            "This will permanently delete vendor \"%s\". This action cannot be undone.",
          ),
          vendor.name,
        ),
      },
    );
  };
};

export const vendorConnectionKey = "VendorsPage_vendors";

export const vendorsQuery = graphql`
  query VendorGraphListQuery($organizationId: ID!, $snapshotId: ID) {
    node(id: $organizationId) {
      ... on Organization {
        id
        canCreateVendor: permission(action: "core:vendor:create")
        ...VendorGraphPaginatedFragment @arguments(snapshotId: $snapshotId)
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
    snapshotId: { type: "ID", defaultValue: null }
  ) {
    vendors(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: { snapshotId: $snapshotId }
    ) @connection(key: "VendorsListQuery_vendors", filters: ["filter"]) {
      __id
      edges {
        node {
          id
          snapshotId
          name
          websiteUrl
          updatedAt
          riskAssessments(
            first: 1
            orderBy: { direction: DESC, field: CREATED_AT }
          ) {
            edges {
              node {
                id
                createdAt
                expiresAt
                dataSensitivity
                businessImpact
              }
            }
          }
          canUpdate: permission(action: "core:vendor:update")
          canDelete: permission(action: "core:vendor:delete")
        }
      }
    }
  }
`;

export const vendorNodeQuery = graphql`
  query VendorGraphNodeQuery($vendorId: ID!) {
    node(id: $vendorId) {
      id
      ... on Vendor {
        snapshotId
        name
        websiteUrl
        canAssess: permission(action: "core:vendor:assess")
        canUpdate: permission(action: "core:vendor:update")
        canDelete: permission(action: "core:vendor:delete")
        canUploadComplianceReport: permission(
          action: "core:vendor-compliance-report:upload"
        )
        canCreateRiskAssessment: permission(
          action: "core:vendor-risk-assessment:create"
        )
        canCreateContact: permission(action: "core:vendor-contact:create")
        canCreateService: permission(action: "core:vendor-service:create")
        canUploadBAA: permission(
          action: "core:vendor-business-associate-agreement:upload"
        )
        canUploadDPA: permission(
          action: "core:vendor-data-privacy-agreement:upload"
        )
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
      id
    }
  }
`;

export const vendorsSelectQuery = graphql`
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
    { fetchPolicy: "network-only" },
  );
  return useMemo(() => {
    return data.organization?.vendors?.edges.map(edge => edge.node) ?? [];
  }, [data]);
}
