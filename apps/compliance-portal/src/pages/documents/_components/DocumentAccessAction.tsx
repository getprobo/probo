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

import { ArrowSquareOutIcon, ClockIcon, LockSimpleIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { useTranslation } from "react-i18next";

interface DocumentAccessActionProps {
  // Whether the viewer may open the document (public or granted access).
  isAuthorized: boolean;
  // Whether an access request is already pending for this document.
  requested: boolean;
  // Opens the document; only invoked when authorized.
  onView: () => void;
  // Whether the export/open is in flight.
  isViewing: boolean;
}

// Trailing access control for a document entry: "View" when authorized, a
// pending label when access was requested, otherwise a (currently inert) "Get
// Access" call to action.
export function DocumentAccessAction({ isAuthorized, requested, onView, isViewing }: DocumentAccessActionProps) {
  const { t } = useTranslation("documents");

  if (isAuthorized) {
    return (
      <Button
        variant="ghost"
        color="neutral"
        highContrast
        iconStart={<ArrowSquareOutIcon />}
        loading={isViewing}
        onClick={onView}
      >
        {t("actions.view")}
      </Button>
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
    <Button variant="ghost" color="neutral" highContrast iconStart={<LockSimpleIcon />}>
      {t("actions.getAccess")}
    </Button>
  );
}
