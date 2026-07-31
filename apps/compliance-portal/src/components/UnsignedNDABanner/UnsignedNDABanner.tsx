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

import { FileTextIcon } from "@phosphor-icons/react";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import { Banner } from "#/components/Banner/Banner";
import { useLocalizedPath } from "#/lib/i18n/useLocale";

import type { UnsignedNDABanner_compliancePortal$key } from "./__generated__/UnsignedNDABanner_compliancePortal.graphql";

const unsignedNDABannerFragment = graphql`
  fragment UnsignedNDABanner_compliancePortal on CompliancePortal {
    nonDisclosureAgreement {
      viewerSignature {
        status
      }
    }
  }
`;

interface UnsignedNDABannerProps {
  compliancePortalKey: UnsignedNDABanner_compliancePortal$key;
}

// Full-bleed notice when a signed-in user still needs to complete the portal
// NDA. Dismissed state is React-only (no localStorage/cookies).
export function UnsignedNDABanner({ compliancePortalKey }: UnsignedNDABannerProps) {
  const { t } = useTranslation();
  const localizedPath = useLocalizedPath();
  const portal = useFragment(unsignedNDABannerFragment, compliancePortalKey);
  const [dismissed, setDismissed] = useState(false);

  const signature = portal.nonDisclosureAgreement?.viewerSignature;
  const needsSignature = signature != null && signature.status !== "COMPLETED";
  const visible = !dismissed && needsSignature;

  if (!visible) {
    return null;
  }

  const ndaHref = `${localizedPath("/nda")}?continue=${encodeURIComponent(window.location.href)}`;

  return (
    <Banner
      color="amber"
      icon={<FileTextIcon weight="fill" />}
      message={(
        <Text size={2} color="neutral" highContrast>
          {t("nda.unsignedBanner.message")}
        </Text>
      )}
      actions={(
        <ButtonLink
          to={ndaHref}
          size={1}
          variant="ghost"
          color="amber"
        >
          {t("nda.unsignedBanner.sign")}
        </ButtonLink>
      )}
      dismissLabel={t("nda.unsignedBanner.dismiss")}
      onDismiss={() => setDismissed(true)}
    />
  );
}
