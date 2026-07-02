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

import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";

import type { MeasureGraphDeleteMutation } from "#/__generated__/core/MeasureGraphDeleteMutation.graphql";

import { useMutationWithToasts } from "../useMutationWithToasts";

export const MeasureConnectionKey = "MeasuresPage_measures";

const deleteMeasureMutation = graphql`
  mutation MeasureGraphDeleteMutation(
    $input: DeleteMeasureInput!
    $connections: [ID!]!
  ) {
    deleteMeasure(input: $input) {
      deletedMeasureId @deleteEdge(connections: $connections)
    }
  }
`;

export function useDeleteMeasureMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<MeasureGraphDeleteMutation>(
    deleteMeasureMutation,
    {
      successMessage: __("Measure deleted successfully."),
      errorMessage: __("Failed to delete measure"),
    },
  );
}

const measureUpdateMutation = graphql`
  mutation MeasureGraphUpdateMutation($input: UpdateMeasureInput!) {
    updateMeasure(input: $input) {
      measure {
        ...MeasureFormDialogMeasureFragment
      }
    }
  }
`;

export const useUpdateMeasure = () => {
  const { __ } = useTranslate();

  return useMutationWithToasts(measureUpdateMutation, {
    successMessage: __("Measure updated successfully."),
    errorMessage: __("Failed to update measure"),
  });
};
