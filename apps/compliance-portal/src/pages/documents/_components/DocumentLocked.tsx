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

import { LockSimpleIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { useTranslation } from "react-i18next";

import { EmptyState } from "#/components/EmptyState/EmptyState";

interface DocumentLockedProps {
  // Requests access for the locked resource (prompting sign-in first when
  // needed).
  onGetAccess: () => void;
  // Whether the access request is in flight.
  isRequesting: boolean;
}

// Shown when the viewer resolves a document the visitor may not access, with a
// Get Access CTA that requests access (prompting sign-in first when needed).
export function DocumentLocked({ onGetAccess, isRequesting }: DocumentLockedProps) {
  const { t } = useTranslation("documents");

  return (
    <div className="grid h-full place-items-center bg-sand-3 p-8">
      <EmptyState
        icon={<LockSimpleIcon />}
        title={t("viewer.locked.title")}
        description={t("viewer.locked.description")}
        action={(
          <Button
            color="neutral"
            highContrast
            loading={isRequesting}
            iconStart={<LockSimpleIcon />}
            onClick={onGetAccess}
          >
            {t("actions.getAccess")}
          </Button>
        )}
      />
    </div>
  );
}
