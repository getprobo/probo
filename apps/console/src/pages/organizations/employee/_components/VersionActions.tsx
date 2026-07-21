// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { Button, Spinner } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { VersionActionsFragment$key } from "#/__generated__/core/VersionActionsFragment.graphql";

const fragment = graphql`
  fragment VersionActionsFragment on EmployeeDocumentVersion {
    id
    signed
    consentText
  }
`;

export function VersionActions({
  fKey,
  isSigning,
  onSign,
  onBack,
}: {
  fKey: VersionActionsFragment$key;
  isSigning: boolean;
  onSign: (versionId: string) => void;
  onBack: () => void;
}) {
  const { t } = useTranslation();
  const versionData = useFragment<VersionActionsFragment$key>(fragment, fKey);
  const isSigned = versionData.signed;

  if (isSigned) {
    return (
      <>
        <Button onClick={onBack} className="h-10 w-full" variant="secondary">
          {t("versionActions.actions.back")}
        </Button>
        <p className="text-xs text-txt-tertiary mt-2 h-5" />
      </>
    );
  }

  return (
    <>
      <Button
        onClick={() => onSign(versionData.id)}
        className="h-10 w-full"
        disabled={isSigning}
        icon={isSigning ? Spinner : undefined}
      >
        {t("versionActions.actions.reviewAndSign")}
      </Button>
      <p className="text-xs text-txt-tertiary mt-2">
        {versionData.consentText}
      </p>
    </>
  );
}
