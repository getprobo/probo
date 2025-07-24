import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import type { TrustCenterGraphUpdateMutation } from "./__generated__/TrustCenterGraphUpdateMutation.graphql";

export const trustCenterQuery = graphql`
  query TrustCenterGraphQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        id
        name
        trustCenter {
          id
          active
          slug
          createdAt
          updatedAt
        }
        documents(first: 100) {
          edges {
            node {
              id
              ...TrustCenterDocumentsCardFragment
            }
          }
        }
        audits(first: 100) {
          edges {
            node {
              id
              ...TrustCenterAuditsCardFragment
            }
          }
        }
        vendors(first: 100) {
          edges {
            node {
              id
              ...TrustCenterVendorsCardFragment
            }
          }
        }
      }
    }
  }
`;

export const updateTrustCenterMutation = graphql`
  mutation TrustCenterGraphUpdateMutation($input: UpdateTrustCenterInput!) {
    updateTrustCenter(input: $input) {
      trustCenter {
        id
        active
        slug
        updatedAt
      }
    }
  }
`;

export function useUpdateTrustCenterMutation() {
  return useMutationWithToasts<TrustCenterGraphUpdateMutation>(
    updateTrustCenterMutation,
    {
      successMessage: "Trust center updated successfully",
      errorMessage: "Failed to update trust center",
    }
  );
}
