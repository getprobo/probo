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
import { ButtonAnchor } from "@probo/ui/src/v2/Button/ButtonAnchor";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { ErrorState } from "@probo/ui/src/v2/ErrorState/ErrorState";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { getSafeContinueUrl, redirectToInitiate } from "#/lib/auth/continueUrl";
import { useLocalizedPath } from "#/lib/i18n/useLocale";
import { usePortalContactEmail } from "#/lib/portal/portalContactContext";
import { NotFoundError } from "#/lib/relay/errors";

type ErrorKind = "notFound" | "forbidden" | "server" | "generic";

interface ErrorContent {
  kind: ErrorKind;
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
    return {
      kind: "notFound",
      code: "404",
      titleKey: "errors.notFound.title",
      descriptionKey: "errors.notFound.description",
    };
  }

  if (
    error instanceof ForbiddenError
    || error instanceof UnAuthenticatedError
    || message.includes("FORBIDDEN")
    || message.includes("UNAUTHENTICATED")
  ) {
    return {
      kind: "forbidden",
      code: "403",
      titleKey: "errors.forbidden.title",
      descriptionKey: "errors.forbidden.description",
    };
  }

  if (error instanceof InternalServerError || message.includes("INTERNAL_SERVER_ERROR")) {
    return {
      kind: "server",
      code: "500",
      titleKey: "errors.serverError.title",
      descriptionKey: "errors.serverError.description",
    };
  }

  return {
    kind: "generic",
    titleKey: "errors.generic.title",
    descriptionKey: "errors.generic.description",
  };
}

interface GlobalErrorProps {
  error: unknown;
  // When provided, a "Try again" action is available (500 / generic).
  onRetry?: () => void;
  // Full viewport (standalone) vs inside the app chrome (in-shell).
  fullPage?: boolean;
}

// Page-level error fallback: renders the v2 ErrorState with portal copy and
// status-specific actions (Figma "Error message / Page").
export function GlobalError({ error, onRetry, fullPage = false }: GlobalErrorProps) {
  const { t } = useTranslation();
  const localizedPath = useLocalizedPath();
  const contactEmail = usePortalContactEmail();
  const { kind, code, titleKey, descriptionKey } = resolveContent(error);

  const contactSupport = contactEmail != null
    ? (
        <ButtonAnchor
          href={`mailto:${contactEmail}`}
          variant="soft"
          color="neutral"
          size={2}
        >
          {t("errors.actions.contactSupport")}
        </ButtonAnchor>
      )
    : null;

  const backHome = (variant: "solid" | "soft") => (
    <ButtonLink
      to={localizedPath("/")}
      variant={variant}
      color="neutral"
      highContrast={variant === "solid"}
      size={2}
    >
      {t("errors.actions.backToCompliancePortal")}
    </ButtonLink>
  );

  const tryAgain = onRetry != null
    ? (
        <Button
          variant="solid"
          color="neutral"
          highContrast
          size={2}
          onClick={onRetry}
        >
          {t("errors.actions.tryAgain")}
        </Button>
      )
    : null;

  let actions: ReactNode;
  switch (kind) {
    case "notFound":
      actions = (
        <>
          {backHome("solid")}
          {contactSupport}
        </>
      );
      break;
    case "forbidden":
      actions = (
        <>
          <Button
            variant="solid"
            color="neutral"
            highContrast
            size={2}
            onClick={() => {
              redirectToInitiate(getSafeContinueUrl(window.location.href));
            }}
          >
            {t("errors.actions.requestAccess")}
          </Button>
          {backHome("soft")}
        </>
      );
      break;
    case "server":
    case "generic":
      actions = (
        <>
          {tryAgain ?? backHome("solid")}
          {contactSupport}
        </>
      );
      break;
  }

  return (
    <ErrorState
      fullPage={fullPage}
      code={code}
      title={t(titleKey)}
      description={t(descriptionKey)}
      actions={actions}
    />
  );
}
