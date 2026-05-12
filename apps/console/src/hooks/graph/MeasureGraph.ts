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
