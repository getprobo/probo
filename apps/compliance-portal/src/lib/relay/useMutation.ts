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
