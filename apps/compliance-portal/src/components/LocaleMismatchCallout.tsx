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

import { GlobeIcon, XIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import {
  isUrlLocale,
  URL_LOCALE_LABELS,
} from "#/lib/i18n/locale";
import { useChangeLocale } from "#/lib/i18n/useChangeLocale";
import { useLocale } from "#/lib/i18n/useLocale";
import { useUpdateLocale } from "#/lib/i18n/useUpdateLocale";

import type { LocaleMismatchCallout_identity$key } from "./__generated__/LocaleMismatchCallout_identity.graphql";

const localeMismatchCalloutFragment = graphql`
  fragment LocaleMismatchCallout_identity on Identity {
    locale
  }
`;

interface LocaleMismatchCalloutProps {
  identityKey: LocaleMismatchCallout_identity$key;
}

// Soft notice when the URL locale differs from the signed-in identity preference.
// Dismissed state is React-only (no localStorage/cookies).
export function LocaleMismatchCallout({ identityKey }: LocaleMismatchCalloutProps) {
  const { t } = useTranslation();
  const identity = useFragment(localeMismatchCalloutFragment, identityKey);
  const urlLocale = useLocale();
  const [changeLocale, isChanging] = useChangeLocale();
  const [updateLocale, isUpdating] = useUpdateLocale();
  const [dismissed, setDismissed] = useState(false);

  const savedLocale = isUrlLocale(identity.locale) ? identity.locale : null;

  if (dismissed || savedLocale == null || savedLocale === urlLocale) {
    return null;
  }

  const urlLabel = URL_LOCALE_LABELS[urlLocale];
  const savedLabel = URL_LOCALE_LABELS[savedLocale];
  const busy = isChanging || isUpdating;

  const switchToSaved = () => {
    void changeLocale(savedLocale, { persist: false });
  };

  const adoptUrlLocale = () => {
    void updateLocale(urlLocale).then(() => setDismissed(true));
  };

  return (
    <div className="border-b border-sand-6 bg-sand-2 px-4 py-3">
      <Callout size={1} variant="soft" color="gold" icon={<GlobeIcon weight="fill" />}>
        <div className="flex flex-wrap items-center gap-3">
          <p className="min-w-0 flex-1 text-2 text-sand-12">
            {t("locale.mismatch.message", { language: urlLabel })}
          </p>
          <div className="flex flex-wrap items-center gap-2">
            <Button
              size={1}
              variant="soft"
              color="neutral"
              disabled={busy}
              onClick={switchToSaved}
            >
              {t("locale.mismatch.switchToMine", { language: savedLabel })}
            </Button>
            <Button
              size={1}
              variant="solid"
              color="neutral"
              highContrast
              disabled={busy}
              onClick={adoptUrlLocale}
            >
              {t("locale.mismatch.useThis", { language: urlLabel })}
            </Button>
            <IconButton
              size={1}
              variant="ghost"
              color="neutral"
              aria-label={t("locale.mismatch.dismiss")}
              disabled={busy}
              onClick={() => setDismissed(true)}
            >
              <XIcon />
            </IconButton>
          </div>
        </div>
      </Callout>
    </div>
  );
}
