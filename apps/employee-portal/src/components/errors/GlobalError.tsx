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

import {
  ForbiddenError,
  InternalServerError,
  MembershipRequiredError,
} from "@probo/relay";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonAnchor } from "@probo/ui/src/v2/Button/ButtonAnchor";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { ErrorState } from "@probo/ui/src/v2/ErrorState/ErrorState";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { NotFoundError } from "#/lib/relay/errors";

type ErrorKind = "notFound" | "forbidden" | "server" | "generic";

interface ErrorContent {
  kind: ErrorKind;
  code?: string;
  titleKey: string;
  descriptionKey: string;
}

const SUPPORT_HREF = "mailto:support@probo.com";

// Map a caught error to the page-level copy. Recognizes the typed error
// classes first, then falls back to the code embedded in generic Error
// messages.
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
    || error instanceof MembershipRequiredError
    || message.includes("FORBIDDEN")
    || message.includes("MEMBERSHIP_REQUIRED")
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

// Page-level error fallback: renders the v2 ErrorState with employee-portal
// copy and status-specific actions.
export function GlobalError({ error, onRetry, fullPage = false }: GlobalErrorProps) {
  const { t } = useTranslation();
  const { kind, code, titleKey, descriptionKey } = resolveContent(error);

  const contactSupport = (
    <ButtonAnchor
      href={SUPPORT_HREF}
      variant="soft"
      color="neutral"
      size={2}
    >
      {t("errors.actions.contactSupport")}
    </ButtonAnchor>
  );

  const backToOrganizations = (variant: "solid" | "soft") => (
    <ButtonLink
      to="/"
      variant={variant}
      color="neutral"
      highContrast={variant === "solid"}
      size={2}
    >
      {t("errors.actions.backToOrganizations")}
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
    case "forbidden":
      actions = (
        <>
          {backToOrganizations("solid")}
          {contactSupport}
        </>
      );
      break;
    case "server":
    case "generic":
      actions = (
        <>
          {tryAgain ?? backToOrganizations("solid")}
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
