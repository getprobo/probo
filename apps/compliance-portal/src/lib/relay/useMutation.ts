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

import { Toast } from "@base-ui/react/toast";
import { formatError, type GraphQLError } from "@probo/helpers";
import { createUseMutation, type MutationNotifier } from "@probo/relay";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";

/**
 * Binds the shared awaitable useMutation (`@probo/relay`) to this app's
 * feedback stack: Base UI toasts, i18next titles, and `formatError`
 * descriptions. This is the only place those opinions are wired.
 *
 * Always import useMutation from `#/lib/relay/useMutation` — never useMutation
 * from react-relay.
 */
function useMutationNotifier(): MutationNotifier {
  const toast = Toast.useToastManager();
  const { t } = useTranslation();

  return useMemo<MutationNotifier>(
    () => ({
      notifySuccess: (title) => {
        toast.add({ title, type: "success" });
      },
      notifyError: (error, title) => {
        const finalTitle = title ?? t("common.error");
        toast.add({
          title: finalTitle,
          description: formatError(finalTitle, error as GraphQLError),
          type: "error",
        });
      },
    }),
    [toast, t],
  );
}

export type { MutationFeedback } from "@probo/relay";

export const useMutation = createUseMutation(useMutationNotifier);
