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

import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { type KeyboardEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { ManualOrgInput_source$key } from "#/__generated__/core/ManualOrgInput_source.graphql";

import { manualOrgInput } from "../sources/_components/variants";

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
  const { field } = manualOrgInput();

  const handleBlur = () => {
    const trimmed = value.trim();
    if (trimmed && trimmed !== selectedOrganization) {
      onSubmit(trimmed);
    }
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Enter") {
      event.preventDefault();
      handleBlur();
    }
  };

  return (
    <Field label={t("accessReviewSourceRow.organizations.empty.manualLabel")}>
      <TextField
        className={field()}
        size={1}
        placeholder={t("accessReviewSourceRow.organizationSlugPlaceholder")}
        value={value}
        onValueChange={setValue}
        onBlur={handleBlur}
        onKeyDown={handleKeyDown}
      />
    </Field>
  );
}
