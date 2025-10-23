import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import type { TrustCenterGraphUpdateMutation } from "./__generated__/TrustCenterGraphUpdateMutation.graphql";

export const trustCenterQuery = graphql`
  query TrustCenterGraphQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        id
        name
        customDomain {
          id
          domain
        }
        trustCenter {
          id
          active
          ndaFileName
          ndaFileUrl
          createdAt
          updatedAt
          references(first: 100, orderBy: { field: CREATED_AT, direction: DESC }) {
            edges {
              node {
                id
                name
                description
                websiteUrl
                logoUrl
                createdAt
                updatedAt
              }
            }
          }
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
        slackConnections(first: 100) {
          edges {
            node {
              id
              channel
              channelId
              createdAt
              updatedAt
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

export const uploadTrustCenterNDAMutation = graphql`
  mutation TrustCenterGraphUploadNDAMutation($input: UploadTrustCenterNDAInput!) {
    uploadTrustCenterNDA(input: $input) {
      trustCenter {
        id
        ndaFileName
        updatedAt
      }
    }
  }
`;

export function useUploadTrustCenterNDAMutation() {
  return useMutationWithToasts(
    uploadTrustCenterNDAMutation,
    {
      successMessage: "NDA uploaded successfully",
      errorMessage: "Failed to upload NDA",
    }
  );
}

export const deleteTrustCenterNDAMutation = graphql`
  mutation TrustCenterGraphDeleteNDAMutation($input: DeleteTrustCenterNDAInput!) {
    deleteTrustCenterNDA(input: $input) {
      trustCenter {
        id
        ndaFileName
        updatedAt
      }
    }
  }
`;

export function useDeleteTrustCenterNDAMutation() {
  return useMutationWithToasts(
    deleteTrustCenterNDAMutation,
    {
      successMessage: "NDA deleted successfully",
      errorMessage: "Failed to delete NDA",
    }
  );
}
