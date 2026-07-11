// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useToast } from "@probo/ui";
import { useMutation } from "react-relay";

import type { accessReviewSourceMutationsCreateMutation } from "#/__generated__/core/accessReviewSourceMutationsCreateMutation.graphql";

import { createAccessReviewSourceMutation } from "../accessReviewSourceMutations";

type UseCreateAccessReviewSourceParams = {
  organizationId: string;
  connectionId: string;
  onSuccess: () => void;
};

// useCreateAccessReviewSource wraps the shared "after a connector is created,
// create the access source, toast the outcome, and close the main dialog on
// success" flow. It returns createSourceAfterConnector, which runs the source
// mutation and invokes onDone on both success and error (for the caller's own
// cleanup) before toasting; onSuccess is called to close the MAIN dialog.
export function useCreateAccessReviewSource({
  organizationId,
  connectionId,
  onSuccess,
}: UseCreateAccessReviewSourceParams) {
  const { __ } = useTranslate();
  const { toast } = useToast();

  const [createAccessReviewSource]
    = useMutation<accessReviewSourceMutationsCreateMutation>(
      createAccessReviewSourceMutation,
    );

  const createSourceAfterConnector = (
    connectorId: string,
    displayName: string,
    onDone: () => void,
  ) => {
    createAccessReviewSource({
      variables: {
        input: {
          organizationId,
          connectorId,
          name: displayName,
          csvData: null,
        },
        connections: [connectionId],
      },
      onCompleted(_, errors) {
        onDone();
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to create access source"),
              errors,
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Access source created successfully."),
          variant: "success",
        });
        onSuccess();
      },
      onError(error) {
        onDone();
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to create access source"),
            error,
          ),
          variant: "error",
        });
      },
    });
  };

  return createSourceAfterConnector;
}
