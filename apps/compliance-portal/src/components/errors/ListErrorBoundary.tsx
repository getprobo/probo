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

import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { InlineError } from "@probo/ui/src/v2/InlineError/InlineError";
import { type ReactNode, useState } from "react";
import { useTranslation } from "react-i18next";

interface ListErrorBoundaryProps {
  // Refetch the list from the network, calling `done` once the request settles.
  // The caller owns the refetch (it holds the refetchable fragment); this keeps
  // that function above the boundary so it survives the child's error.
  onRetry: (done: () => void) => void;
  children: ReactNode;
}

// Contains a list field error to an inline fallback with a working retry. A
// standalone list page has no section framing, so the fallback is a bare
// InlineError (no card). The boundary only resets *after* the caller's refetch
// settles (via the `done` callback bumping its key), so remounting reads the
// refreshed store instead of racing the in-flight request back into the same
// error. See contrib/claude/error-handling.md.
export function ListErrorBoundary({ onRetry, children }: ListErrorBoundaryProps) {
  const { t } = useTranslation();
  const [resetToken, setResetToken] = useState(0);

  return (
    <ErrorBoundary
      key={resetToken}
      fallback={(
        <div className="py-8">
          <InlineError
            message={t("errors.inline.message")}
            retryLabel={t("errors.inline.retry")}
            onRetry={() => onRetry(() => setResetToken(token => token + 1))}
          />
        </div>
      )}
    >
      {children}
    </ErrorBoundary>
  );
}
