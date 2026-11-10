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

import { CaretDownIcon, PencilSimpleIcon } from "@phosphor-icons/react";
import { Button as ButtonV2 } from "@probo/ui/src/v2/Button/Button";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownRadioGroup } from "@probo/ui/src/v2/Dropdown/DropdownRadioGroup";
import { DropdownRadioItem } from "@probo/ui/src/v2/Dropdown/DropdownRadioItem";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { useTranslation } from "react-i18next";

export type DocumentReviewMode = "viewing" | "editing" | "suggesting";

interface DocumentReviewModeMenuProps {
  mode: DocumentReviewMode;
  canEdit: boolean;
  canSuggest: boolean;
  onModeChange: (mode: DocumentReviewMode) => void;
}

export function DocumentReviewModeMenu({
  mode,
  canEdit,
  canSuggest,
  onModeChange,
}: DocumentReviewModeMenuProps) {
  const { t } = useTranslation();

  return (
    <Dropdown>
      <DropdownTrigger
        render={(
          <ButtonV2
            variant="ghost"
            color={mode === "suggesting" ? "green" : "neutral"}
            size={2}
            iconStart={<PencilSimpleIcon />}
            iconEnd={<CaretDownIcon />}
          >
            {mode === "viewing"
              ? t("documentDescriptionPage.review.viewing")
              : mode === "suggesting"
                ? t("documentDescriptionPage.review.suggesting")
                : t("documentDescriptionPage.review.editing")}
          </ButtonV2>
        )}
      />
      <DropdownPopup align="end">
        <DropdownRadioGroup
          value={mode}
          onValueChange={(value: string) => {
            if (value === "viewing" || value === "editing" || value === "suggesting") {
              onModeChange(value);
            }
          }}
        >
          <DropdownRadioItem value="viewing">
            {t("documentDescriptionPage.review.viewing")}
          </DropdownRadioItem>
          {canEdit && (
            <DropdownRadioItem value="editing">
              {t("documentDescriptionPage.review.editing")}
            </DropdownRadioItem>
          )}
          {canSuggest && (
            <DropdownRadioItem value="suggesting">
              {t("documentDescriptionPage.review.suggesting")}
            </DropdownRadioItem>
          )}
        </DropdownRadioGroup>
      </DropdownPopup>
    </Dropdown>
  );
}
