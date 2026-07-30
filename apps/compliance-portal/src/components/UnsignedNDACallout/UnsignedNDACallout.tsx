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

import { FileTextIcon, XIcon } from "@phosphor-icons/react";
import { Link } from "@probo/ui/src/v2/Button/Link";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import { useLocalizedPath } from "#/lib/i18n/useLocale";

import type { UnsignedNDACallout_compliancePortal$key } from "./__generated__/UnsignedNDACallout_compliancePortal.graphql";
import { unsignedNDACallout } from "./variants";

const unsignedNDACalloutFragment = graphql`
  fragment UnsignedNDACallout_compliancePortal on CompliancePortal {
    nonDisclosureAgreement {
      viewerSignature {
        status
      }
    }
  }
`;

interface UnsignedNDACalloutProps {
  compliancePortalKey: UnsignedNDACallout_compliancePortal$key;
}

// Full-bleed notice when a signed-in user still needs to complete the portal
// NDA. Dismissed state is React-only (no localStorage/cookies).
export function UnsignedNDACallout({ compliancePortalKey }: UnsignedNDACalloutProps) {
  const { t } = useTranslation();
  const localizedPath = useLocalizedPath();
  const portal = useFragment(unsignedNDACalloutFragment, compliancePortalKey);
  const [dismissed, setDismissed] = useState(false);

  const signature = portal.nonDisclosureAgreement?.viewerSignature;
  const needsSignature = signature != null && signature.status !== "COMPLETED";
  const visible = !dismissed && needsSignature;

  if (!visible) {
    return null;
  }

  const ndaHref = `${localizedPath("/nda")}?continue=${encodeURIComponent(window.location.href)}`;
  const slots = unsignedNDACallout();

  return (
    <aside className={slots.root()} role="status">
      <div className={slots.content()}>
        <FileTextIcon weight="fill" className={slots.icon()} aria-hidden />
        <Text size={2} color="neutral" highContrast className={slots.message()}>
          {t("nda.unsignedBanner.message")}
        </Text>
        <IconButton
          size={1}
          variant="ghost"
          color="neutral"
          aria-label={t("nda.unsignedBanner.dismiss")}
          className={slots.dismissMobile()}
          onClick={() => setDismissed(true)}
        >
          <XIcon />
        </IconButton>
      </div>
      <div className={slots.actions()}>
        <Link
          to={ndaHref}
          size={1}
          variant="solid"
          color="gold"
          className={slots.cta()}
        >
          {t("nda.unsignedBanner.sign")}
        </Link>
        <IconButton
          size={1}
          variant="ghost"
          color="neutral"
          aria-label={t("nda.unsignedBanner.dismiss")}
          className={slots.dismissDesktop()}
          onClick={() => setDismissed(true)}
        >
          <XIcon />
        </IconButton>
      </div>
    </aside>
  );
}
