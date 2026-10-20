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

import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import type { InternalControlGraphDeleteMutation } from "#/__generated__/core/InternalControlGraphDeleteMutation.graphql";
import type { InternalControlGraphUpdateMutation } from "#/__generated__/core/InternalControlGraphUpdateMutation.graphql";

import { useMutationWithToasts } from "../useMutationWithToasts";

export const InternalControlConnectionKey = "InternalControlsPage_internalControls";

const deleteInternalControlMutation = graphql`
  mutation InternalControlGraphDeleteMutation(
    $input: DeleteInternalControlInput!
    $connections: [ID!]!
  ) {
    deleteInternalControl(input: $input) {
      deletedInternalControlId @deleteEdge(connections: $connections)
    }
  }
`;

export function useDeleteInternalControlMutation() {
  const { t } = useTranslation();

  return useMutationWithToasts<InternalControlGraphDeleteMutation>(
    deleteInternalControlMutation,
    {
      successMessage: t("internalControlGraph.messages.deleted"),
      errorMessage: t("internalControlGraph.errors.delete"),
    },
  );
}

const internalControlUpdateMutation = graphql`
  mutation InternalControlGraphUpdateMutation($input: UpdateInternalControlInput!) {
    updateInternalControl(input: $input) {
      internalControl {
        ...InternalControlFormDialogInternalControlFragment
        nextEvidenceDue
        nextTestDue
      }
    }
  }
`;

export const useUpdateInternalControl = () => {
  const { t } = useTranslation();

  return useMutationWithToasts<InternalControlGraphUpdateMutation>(internalControlUpdateMutation, {
    successMessage: t("internalControlGraph.messages.updated"),
    errorMessage: t("internalControlGraph.errors.update"),
  });
};
