// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { formatError } from "@probo/helpers";
import { useToast } from "@probo/ui";
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();
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
            title: t("useCreateAccessReviewSource.messages.error"),
            description: formatError(
              t("useCreateAccessReviewSource.errors.create"),
              errors,
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("useCreateAccessReviewSource.messages.success"),
          description: t("useCreateAccessReviewSource.messages.created"),
          variant: "success",
        });
        onSuccess();
      },
      onError(error) {
        onDone();
        toast({
          title: t("useCreateAccessReviewSource.messages.error"),
          description: formatError(
            t("useCreateAccessReviewSource.errors.create"),
            error,
          ),
          variant: "error",
        });
      },
    });
  };

  return createSourceAfterConnector;
}
