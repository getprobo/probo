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
        nda {
          fileName
          downloadUrl
        }
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
        nda {
          fileName
          downloadUrl
        }
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
