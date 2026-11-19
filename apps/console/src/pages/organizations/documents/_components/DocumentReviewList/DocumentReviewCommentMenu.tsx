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

import { DotsThreeVerticalIcon, PencilSimpleIcon, TrashIcon } from "@phosphor-icons/react";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { useTranslation } from "react-i18next";

interface DocumentReviewCommentMenuProps {
  canEdit: boolean;
  canDelete: boolean;
  onEdit: () => void;
  onDelete: () => void;
}

export function DocumentReviewCommentMenu({
  canEdit,
  canDelete,
  onEdit,
  onDelete,
}: DocumentReviewCommentMenuProps) {
  const { t } = useTranslation("organizations/documents");

  if (!canEdit && !canDelete) {
    return null;
  }

  return (
    <Dropdown>
      <Tooltip>
        <TooltipTrigger
          render={(
            <DropdownTrigger
              render={(
                <IconButton
                  size={1}
                  variant="ghost"
                  color="neutral"
                  aria-label={t("details.review.more")}
                />
              )}
            />
          )}
        >
          <DotsThreeVerticalIcon />
        </TooltipTrigger>
        <TooltipPopup>{t("details.review.more")}</TooltipPopup>
      </Tooltip>
      <DropdownPopup align="end">
        {canEdit && (
          <DropdownItem color="neutral" iconStart={<PencilSimpleIcon />} onClick={onEdit}>
            {t("details.review.edit")}
          </DropdownItem>
        )}
        {canDelete && (
          <DropdownItem color="error" iconStart={<TrashIcon />} onClick={onDelete}>
            {t("details.review.delete")}
          </DropdownItem>
        )}
      </DropdownPopup>
    </Dropdown>
  );
}
