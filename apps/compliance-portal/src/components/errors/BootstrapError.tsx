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

import { Anchor } from "@probo/ui/src/v2/Button/Anchor";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ErrorState } from "@probo/ui/src/v2/ErrorState/ErrorState";
import { useTranslation } from "react-i18next";

import { getPathPrefix } from "#/lib/http/pathPrefix";

// Outermost fallback for failures that happen before (or in) the router itself.
// It cannot use the router (no context yet), so navigation is a plain anchor and
// recovery is a hard reload.
export function BootstrapError() {
  const { t } = useTranslation();

  return (
    <ErrorState
      fullPage
      title={t("errors.generic.title")}
      description={t("errors.generic.description")}
      actions={(
        <>
          <Anchor href={getPathPrefix() || "/"} variant="solid" color="neutral" highContrast size={2}>
            {t("errors.actions.backToTrustCenter")}
          </Anchor>
          <Button variant="soft" color="neutral" size={2} onClick={() => window.location.reload()}>
            {t("errors.actions.tryAgain")}
          </Button>
        </>
      )}
    />
  );
}
