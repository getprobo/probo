import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";

export const createTrustCenterFileMutation = graphql`
  mutation TrustCenterFileGraphCreateMutation(
    $input: CreateTrustCenterFileInput!
    $connections: [ID!]!
  ) {
    createTrustCenterFile(input: $input) {
      trustCenterFileEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          category
          fileUrl
          trustCenterVisibility
          createdAt
          updatedAt
        }
      }
    }
  }
`;

export function useCreateTrustCenterFileMutation() {
  return useMutationWithToasts(
    createTrustCenterFileMutation,
    {
      successMessage: "File uploaded successfully",
      errorMessage: "Failed to upload file",
    },
  );
}

export const updateTrustCenterFileMutation = graphql`
  mutation TrustCenterFileGraphUpdateMutation($input: UpdateTrustCenterFileInput!) {
    updateTrustCenterFile(input: $input) {
      trustCenterFile {
        id
        name
        category
        trustCenterVisibility
        updatedAt
      }
    }
  }
`;

export function useUpdateTrustCenterFileMutation() {
  return useMutationWithToasts(
    updateTrustCenterFileMutation,
    {
      successMessage: "File updated successfully",
      errorMessage: "Failed to update file",
    },
  );
}

export const deleteTrustCenterFileMutation = graphql`
  mutation TrustCenterFileGraphDeleteMutation(
    $input: DeleteTrustCenterFileInput!
    $connections: [ID!]!
  ) {
    deleteTrustCenterFile(input: $input) {
      deletedTrustCenterFileId @deleteEdge(connections: $connections)
    }
  }
`;

export function useDeleteTrustCenterFileMutation() {
  return useMutationWithToasts(
    deleteTrustCenterFileMutation,
    {
      successMessage: "File deleted successfully",
      errorMessage: "Failed to delete file",
    },
  );
}
