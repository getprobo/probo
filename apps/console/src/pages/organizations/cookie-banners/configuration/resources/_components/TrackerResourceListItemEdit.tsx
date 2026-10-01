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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { trackerResourceListItem } from "../../../variants";

interface TrackerResourceListItemEditProps {
  displayName: string;
  description: string;
  isUpdating: boolean;
  onSave: (data: { displayName: string; description: string }) => void;
  onCancel: () => void;
}

export function TrackerResourceListItemEdit({
  displayName,
  description,
  isUpdating,
  onSave,
  onCancel,
}: TrackerResourceListItemEditProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const [displayNameValue, setDisplayNameValue] = useState(displayName);
  const [descriptionValue, setDescriptionValue] = useState(description);
  const { edit, editField } = trackerResourceListItem();

  function handleSave() {
    onSave({
      displayName: displayNameValue,
      description: descriptionValue,
    });
  }

  return (
    <TableRow>
      <TableCell colSpan={6}>
        <div className={edit()}>
          <div className={editField()}>
            <TextField
              value={displayNameValue}
              onValueChange={setDisplayNameValue}
              placeholder={t("trackerResourceRowEdit.fields.displayNamePlaceholder")}
              aria-label={t("trackerResourceRowEdit.fields.displayNamePlaceholder")}
            />
          </div>
          <div className={editField()}>
            <TextField
              value={descriptionValue}
              onValueChange={setDescriptionValue}
              placeholder={t("trackerResourceRowEdit.fields.descriptionPlaceholder")}
              aria-label={t("trackerResourceRowEdit.fields.descriptionPlaceholder")}
            />
          </div>
          <Button
            variant="solid"
            loading={isUpdating}
            onClick={handleSave}
          >
            {t("trackerResourceRowEdit.actions.save")}
          </Button>
          <Button
            variant="soft"
            color="neutral"
            onClick={onCancel}
          >
            {t("trackerResourceRowEdit.actions.cancel")}
          </Button>
        </div>
      </TableCell>
    </TableRow>
  );
}
