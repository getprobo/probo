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

import { GlobeIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useDeferredValue, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import { Banner } from "#/components/Banner/Banner";
import { DEFAULT_NAMESPACE } from "#/lib/i18n/backend";
import {
  isUrlLocale,
  URL_LOCALE_LABELS,
  urlLocaleToLanguage,
} from "#/lib/i18n/locale";
import { useChangeLocale } from "#/lib/i18n/useChangeLocale";
import { useLocale } from "#/lib/i18n/useLocale";
import { useUpdateLocale } from "#/lib/i18n/useUpdateLocale";

import type { LocaleMismatchBanner_identity$key } from "./__generated__/LocaleMismatchBanner_identity.graphql";

const localeMismatchBannerFragment = graphql`
  fragment LocaleMismatchBanner_identity on Identity {
    locale
  }
`;

interface LocaleMismatchBannerProps {
  identityKey: LocaleMismatchBanner_identity$key;
}

// Full-bleed notice when the URL locale differs from the signed-in identity
// preference. Dismissed state is React-only (no localStorage/cookies).
export function LocaleMismatchBanner({ identityKey }: LocaleMismatchBannerProps) {
  const { t, i18n } = useTranslation();
  const identity = useFragment(localeMismatchBannerFragment, identityKey);
  const urlLocale = useLocale();
  const [changeLocale, isChanging] = useChangeLocale();
  const [updateLocale, isUpdating] = useUpdateLocale();
  const [dismissed, setDismissed] = useState(false);
  // Bumps after the identity-locale catalog loads so the switch button can
  // re-render in that language (it may not be the active i18n language).
  const [, setSavedCatalogTick] = useState(0);

  const savedLocale = isUrlLocale(identity.locale) ? identity.locale : null;
  const savedLanguage = savedLocale != null ? urlLocaleToLanguage(savedLocale) : null;
  const mismatched = savedLocale != null && savedLocale !== urlLocale;
  // Lag the mismatch flag so a transient desync during startTransition locale
  // switches never paints the banner; a real mismatch still shows once settled.
  const deferredMismatched = useDeferredValue(mismatched);
  const visible = !dismissed && mismatched && deferredMismatched;

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
    void updateLocale(urlLocale)
      .then(() => setDismissed(true))
      .catch(() => {});
  };

  return (
    <Banner
      color="sky"
      icon={<GlobeIcon weight="fill" />}
      message={(
        <Text size={2} color="neutral" highContrast>
          {t("locale.mismatch.message", { language: urlLabel })}
        </Text>
      )}
      actions={(
        <>
          <Button
            size={1}
            variant="ghost"
            color="sky"
            disabled={busy}
            onClick={adoptUrlLocale}
          >
            {t("locale.mismatch.useThis", { language: urlLabel })}
          </Button>
          <Button
            size={1}
            variant="solid"
            color="neutral"
            highContrast
            disabled={busy}
            onClick={switchToSaved}
          >
            {t("locale.mismatch.switchToMine", {
              language: savedLabel,
              // Label this action in the user's saved locale so it reads as
              // "switch back to my language", not the page they're visiting.
              lng: savedLanguage,
            })}
          </Button>
        </>
      )}
      dismissLabel={t("locale.mismatch.dismiss")}
      onDismiss={() => setDismissed(true)}
      dismissDisabled={busy}
    />
  );
}
