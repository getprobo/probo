// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { graphql } from "relay-runtime";

import type { TrustCenterGraphUpdateMutation } from "#/__generated__/core/TrustCenterGraphUpdateMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const updateTrustCenterMutation = graphql`
  mutation TrustCenterGraphUpdateMutation($input: UpdateTrustCenterInput!) {
    updateTrustCenter(input: $input) {
      trustCenter {
        id
        active
        searchEngineIndexing
        updatedAt
      }
    }
  }
`;

export function useUpdateTrustCenterMutation() {
  return useMutationWithToasts<TrustCenterGraphUpdateMutation>(
    updateTrustCenterMutation,
    {
      successMessage: "Compliance Page updated successfully",
      errorMessage: "Failed to update compliance page",
    },
  );
}

const uploadTrustCenterNDAMutation = graphql`
  mutation TrustCenterGraphUploadNDAMutation(
    $input: UploadTrustCenterNDAInput!
  ) {
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
  return useMutationWithToasts(uploadTrustCenterNDAMutation, {
    successMessage: "NDA uploaded successfully",
    errorMessage: "Failed to upload NDA",
  });
}

const deleteTrustCenterNDAMutation = graphql`
  mutation TrustCenterGraphDeleteNDAMutation(
    $input: DeleteTrustCenterNDAInput!
  ) {
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
  return useMutationWithToasts(deleteTrustCenterNDAMutation, {
    successMessage: "NDA deleted successfully",
    errorMessage: "Failed to delete NDA",
  });
}
