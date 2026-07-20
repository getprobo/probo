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

import { ArrowRightIcon, ClockIcon, LockSimpleIcon, SpinnerGapIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Link } from "@probo/ui/src/v2/Button/Link";
import { useTranslation } from "react-i18next";

interface DocumentAccessActionProps {
  // Whether the viewer may open the document (public or granted access).
  isAuthorized: boolean;
  // Whether an access request is already pending for this document.
  requested: boolean;
  // Route to the document viewer, used when authorized.
  viewHref: string;
  // Requests access for this entry (gated behind sign-in when needed).
  onGetAccess: () => void;
  // Whether the access request is in flight.
  isRequesting: boolean;
  // When false, render a non-interactive status icon (used as the mobile
  // affordance while the row itself handles the hit target).
  interactive?: boolean;
}

function StatusIcon({
  isAuthorized,
  requested,
  isRequesting,
}: {
  isAuthorized: boolean;
  requested: boolean;
  isRequesting: boolean;
}) {
  if (isAuthorized) {
    return <ArrowRightIcon className="size-4 text-sand-12" />;
  }
  if (requested) {
    return <ClockIcon className="size-4 text-sand-11" />;
  }
  if (isRequesting) {
    return <SpinnerGapIcon className="size-4 animate-spin text-sand-12" />;
  }
  return <LockSimpleIcon className="size-4 text-sand-12" />;
}

// Trailing access control for a document entry: a "View" link to the viewer when
// authorized, a pending label when access was requested, otherwise a "Get
// Access" action that requests access (prompting sign-in first when needed).
export function DocumentAccessAction({
  isAuthorized,
  requested,
  viewHref,
  onGetAccess,
  isRequesting,
  interactive = true,
}: DocumentAccessActionProps) {
  const { t } = useTranslation("documents");

  if (!interactive) {
    // Pending rows have no mobile hit overlay — expose status for assistive tech.
    if (requested) {
      return (
        <span
          className="flex size-8 items-center justify-center"
          role="status"
          aria-label={t("actions.requested")}
        >
          <ClockIcon className="size-4 text-sand-11" aria-hidden />
        </span>
      );
    }

    return (
      <span className="flex size-8 items-center justify-center" aria-hidden>
        <StatusIcon
          isAuthorized={isAuthorized}
          requested={requested}
          isRequesting={isRequesting}
        />
      </span>
    );
  }

  if (isAuthorized) {
    return (
      <Link
        to={viewHref}
        variant="ghost"
        color="neutral"
        highContrast
        iconStart={<ArrowRightIcon />}
      >
        {t("actions.view")}
      </Link>
    );
  }

  if (requested) {
    return (
      <Button variant="ghost" color="neutral" disabled iconStart={<ClockIcon />}>
        {t("actions.requested")}
      </Button>
    );
  }

  return (
    <Button
      variant="ghost"
      color="neutral"
      highContrast
      loading={isRequesting}
      iconStart={<LockSimpleIcon />}
      onClick={onGetAccess}
    >
      {t("actions.getAccess")}
    </Button>
  );
}
