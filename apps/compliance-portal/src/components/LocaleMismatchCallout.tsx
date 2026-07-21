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
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import { DEFAULT_NAMESPACE } from "#/lib/i18n/backend";
import {
  isUrlLocale,
  URL_LOCALE_LABELS,
  urlLocaleToLanguage,
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

// Full-bleed notice when the URL locale differs from the signed-in identity
// preference. Dismissed state is React-only (no localStorage/cookies).
export function LocaleMismatchCallout({ identityKey }: LocaleMismatchCalloutProps) {
  const { t, i18n } = useTranslation();
  const identity = useFragment(localeMismatchCalloutFragment, identityKey);
  const urlLocale = useLocale();
  const [changeLocale, isChanging] = useChangeLocale();
  const [updateLocale, isUpdating] = useUpdateLocale();
  const [dismissed, setDismissed] = useState(false);
  // Bumps after the identity-locale catalog loads so the switch button can
  // re-render in that language (it may not be the active i18n language).
  const [, setSavedCatalogTick] = useState(0);

  const savedLocale = isUrlLocale(identity.locale) ? identity.locale : null;
  const savedLanguage = savedLocale != null ? urlLocaleToLanguage(savedLocale) : null;
  const visible = !dismissed && savedLocale != null && savedLocale !== urlLocale;

  useEffect(() => {
    if (!visible || savedLanguage == null) {
      return;
    }
    if (i18n.hasResourceBundle(savedLanguage, DEFAULT_NAMESPACE)) {
      return;
    }
    let cancelled = false;
    void i18n.loadLanguages(savedLanguage).then(() => {
      if (!cancelled) {
        setSavedCatalogTick(tick => tick + 1);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [visible, savedLanguage, i18n]);

  if (!visible || savedLocale == null || savedLanguage == null) {
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
    <aside
      className="flex w-full items-center gap-3 bg-gold-3 px-8 py-2.5 max-md:flex-col max-md:items-stretch max-md:gap-3 max-md:px-4 max-md:py-3"
      role="status"
    >
      <div className="flex min-w-0 flex-1 items-start gap-2">
        <GlobeIcon weight="fill" className="mt-0.5 size-4 shrink-0 text-gold-11" aria-hidden />
        <Text size={2} color="neutral" highContrast className="min-w-0 flex-1">
          {t("locale.mismatch.message", { language: urlLabel })}
        </Text>
        <IconButton
          size={1}
          variant="ghost"
          color="neutral"
          aria-label={t("locale.mismatch.dismiss")}
          disabled={busy}
          className="md:hidden"
          onClick={() => setDismissed(true)}
        >
          <XIcon />
        </IconButton>
      </div>
      <div className="flex shrink-0 items-center gap-2 max-md:flex-col max-md:items-stretch">
        <Button
          size={1}
          variant="solid"
          color="gold"
          disabled={busy}
          className="max-md:w-full"
          onClick={switchToSaved}
        >
          {t("locale.mismatch.switchToMine", {
            language: savedLabel,
            // Label this action in the user's saved locale so it reads as
            // "switch back to my language", not the page they're visiting.
            lng: savedLanguage,
          })}
        </Button>
        <Button
          size={1}
          variant="solid"
          color="neutral"
          highContrast
          disabled={busy}
          className="max-md:w-full"
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
          className="max-md:hidden"
          onClick={() => setDismissed(true)}
        >
          <XIcon />
        </IconButton>
      </div>
    </aside>
  );
}
