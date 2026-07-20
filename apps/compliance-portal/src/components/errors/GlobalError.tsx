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

import { ForbiddenError, InternalServerError, UnAuthenticatedError } from "@probo/relay";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Link } from "@probo/ui/src/v2/Button/Link";
import { ErrorState } from "@probo/ui/src/v2/ErrorState/ErrorState";
import { useTranslation } from "react-i18next";

import { NotFoundError } from "#/lib/relay/errors";

interface ErrorContent {
  code?: string;
  titleKey: string;
  descriptionKey: string;
}

// Map a caught error to the page-level copy. Recognizes the portal error
// classes first, then falls back to the code embedded in generic Error messages
// (thrown request-level by lib/relay/fetch.ts).
function resolveContent(error: unknown): ErrorContent {
  const message = error instanceof Error ? error.message : "";

  if (error instanceof NotFoundError || message.includes("NOT_FOUND")) {
    return { code: "404", titleKey: "errors.notFound.title", descriptionKey: "errors.notFound.description" };
  }

  if (
    error instanceof ForbiddenError
    || error instanceof UnAuthenticatedError
    || message.includes("FORBIDDEN")
    || message.includes("UNAUTHENTICATED")
  ) {
    return { code: "403", titleKey: "errors.forbidden.title", descriptionKey: "errors.forbidden.description" };
  }

  if (error instanceof InternalServerError || message.includes("INTERNAL_SERVER_ERROR")) {
    return { code: "500", titleKey: "errors.serverError.title", descriptionKey: "errors.serverError.description" };
  }

  return { titleKey: "errors.generic.title", descriptionKey: "errors.generic.description" };
}

interface GlobalErrorProps {
  error: unknown;
  // When provided, a "Try again" secondary action is shown.
  onRetry?: () => void;
  // Full viewport (standalone) vs inside the app chrome (in-shell).
  fullPage?: boolean;
}

// Page-level error fallback: renders the v2 ErrorState with portal copy and
// actions. Used by the route boundaries (root + page).
export function GlobalError({ error, onRetry, fullPage = false }: GlobalErrorProps) {
  const { t } = useTranslation();
  const { code, titleKey, descriptionKey } = resolveContent(error);

  return (
    <ErrorState
      fullPage={fullPage}
      code={code}
      title={t(titleKey)}
      description={t(descriptionKey)}
      actions={(
        <>
          <Link to="/" variant="solid" color="neutral" highContrast size={2}>
            {t("errors.actions.backToCompliancePortal")}
          </Link>
          {onRetry && (
            <Button variant="soft" color="neutral" size={2} onClick={onRetry}>
              {t("errors.actions.tryAgain")}
            </Button>
          )}
        </>
      )}
    />
  );
}
