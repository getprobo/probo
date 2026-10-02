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

import { Input } from "@probo/ui";
import { type KeyboardEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { ManualOrgInput_source$key } from "#/__generated__/core/ManualOrgInput_source.graphql";

const manualOrgInputFragment = graphql`
  fragment ManualOrgInput_source on AccessReviewSource {
    selectedOrganization
  }
`;

interface ManualOrgInputProps {
  sourceKey: ManualOrgInput_source$key;
  onSubmit: (slug: string) => void;
}

export function ManualOrgInput({ sourceKey, onSubmit }: ManualOrgInputProps) {
  const { t } = useTranslation();
  const source = useFragment(manualOrgInputFragment, sourceKey);
  const selectedOrganization = source.selectedOrganization ?? "";
  const [value, setValue] = useState(selectedOrganization);

  const handleBlur = () => {
    const trimmed = value.trim();
    if (trimmed && trimmed !== selectedOrganization) {
      onSubmit(trimmed);
    }
  };

  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleBlur();
    }
  };

  return (
    <Input
      placeholder={t("accessReviewSourceRow.organizationSlugPlaceholder")}
      value={value}
      onChange={e => setValue(e.target.value)}
      onBlur={handleBlur}
      onKeyDown={handleKeyDown}
      className="max-w-40"
    />
  );
}
