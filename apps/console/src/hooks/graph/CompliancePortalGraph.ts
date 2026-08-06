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

import type { CompliancePortalGraphDeleteNDAMutation } from "#/__generated__/core/CompliancePortalGraphDeleteNDAMutation.graphql";
import type { CompliancePortalGraphUpdateMutation } from "#/__generated__/core/CompliancePortalGraphUpdateMutation.graphql";
import type { CompliancePortalGraphUploadNDAMutation } from "#/__generated__/core/CompliancePortalGraphUploadNDAMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const updateCompliancePortalMutation = graphql`
  mutation CompliancePortalGraphUpdateMutation($input: UpdateCompliancePortalInput!) {
    updateCompliancePortal(input: $input) {
      compliancePortal {
        id
        active
        searchEngineIndexing
        rightsRequestsEnabled
        entityName
        description
        websiteUrl
        email
        headquarterAddress
        updatedAt
      }
    }
  }
`;

export function useUpdateCompliancePortalMutation() {
  return useMutation<CompliancePortalGraphUpdateMutation>(
    updateCompliancePortalMutation,
    {
      successMessage: "Compliance portal updated successfully",
      errorToast: "Failed to update compliance portal",
    },
  );
}

const uploadCompliancePortalNDAMutation = graphql`
  mutation CompliancePortalGraphUploadNDAMutation(
    $input: UploadCompliancePortalNDAInput!
  ) {
    uploadCompliancePortalNDA(input: $input) {
      compliancePortal {
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

export function useUploadCompliancePortalNDAMutation() {
  return useMutation<CompliancePortalGraphUploadNDAMutation>(uploadCompliancePortalNDAMutation, {
    successMessage: "NDA uploaded successfully",
    errorToast: "Failed to upload NDA",
  });
}

const deleteCompliancePortalNDAMutation = graphql`
  mutation CompliancePortalGraphDeleteNDAMutation(
    $input: DeleteCompliancePortalNDAInput!
  ) {
    deleteCompliancePortalNDA(input: $input) {
      compliancePortal {
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

export function useDeleteCompliancePortalNDAMutation() {
  return useMutation<CompliancePortalGraphDeleteNDAMutation>(deleteCompliancePortalNDAMutation, {
    successMessage: "NDA deleted successfully",
    errorToast: "Failed to delete NDA",
  });
}
