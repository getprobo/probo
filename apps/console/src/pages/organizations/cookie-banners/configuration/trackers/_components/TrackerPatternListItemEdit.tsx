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

import { fromMaxAgeSeconds, toMaxAgeSeconds } from "@probo/helpers";
import { DurationInput } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { trackerPatternListItem } from "../../../variants";

interface TrackerPatternListItemEditProps {
  pattern: string;
  description: string;
  maxAgeSeconds: number | null;
  isUpdating: boolean;
  onSave: (data: { description: string; maxAgeSeconds: number | null }) => void;
  onCancel: () => void;
}

export function TrackerPatternListItemEdit({
  pattern,
  description,
  maxAgeSeconds,
  isUpdating,
  onSave,
  onCancel,
}: TrackerPatternListItemEditProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const initial = fromMaxAgeSeconds(maxAgeSeconds);
  const [descriptionValue, setDescriptionValue] = useState(description);
  const [duration, setDuration] = useState(initial);
  const { edit, editFields, editField } = trackerPatternListItem();

  function handleSave() {
    onSave({
      description: descriptionValue,
      maxAgeSeconds: toMaxAgeSeconds(duration.value, duration.unit),
    });
  }

  return (
    <TableRow>
      <TableCell colSpan={7}>
        <div className={edit()}>
          <Text size={2} weight="medium" highContrast>
            {pattern}
          </Text>
          <div className={editFields()}>
            <div className={editField()}>
              <Text size={1} color="faint">
                {t("trackerPatternRowEdit.fields.description")}
              </Text>
              <TextField
                value={descriptionValue}
                onValueChange={setDescriptionValue}
                placeholder={t("trackerPatternRowEdit.fields.descriptionPlaceholder")}
                aria-label={t("trackerPatternRowEdit.fields.description")}
              />
            </div>
            <div className={editField()}>
              <Text size={1} color="faint">
                {t("trackerPatternRowEdit.fields.maxAge")}
              </Text>
              <DurationInput
                value={duration.value}
                unit={duration.unit}
                onValueChange={value => setDuration(current => ({ ...current, value }))}
                onUnitChange={unit => setDuration(current => ({ ...current, unit }))}
              />
            </div>
            <Button
              variant="solid"
              loading={isUpdating}
              onClick={handleSave}
            >
              {t("trackerPatternRowEdit.actions.save")}
            </Button>
            <Button
              variant="soft"
              color="neutral"
              onClick={onCancel}
            >
              {t("trackerPatternRowEdit.actions.cancel")}
            </Button>
          </div>
        </div>
      </TableCell>
    </TableRow>
  );
}
