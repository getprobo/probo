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

import { PlusIcon, TrashIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import { userEmailsField } from "../variants";

interface UserEmailsFieldProps {
  value: string[];
  disabled?: boolean;
  readOnly?: boolean;
  onValueChange: (value: string[]) => void;
  onBlur?: () => void;
}

export function UserEmailsField({
  value,
  disabled = false,
  readOnly = false,
  onValueChange,
  onBlur,
}: UserEmailsFieldProps) {
  const { t } = useTranslation();
  const { root, row, field } = userEmailsField();
  const emails = value.length > 0 ? value : readOnly ? [] : [""];

  if (readOnly) {
    if (value.length === 0) {
      return <Text size={2} color="faint">{t("userPage.empty")}</Text>;
    }
    return (
      <div className={root()}>
        {value.map(email => (
          <Text key={email} size={2}>{email}</Text>
        ))}
      </div>
    );
  }

  return (
    <div className={root()}>
      {emails.map((email, index) => (
        <div key={index} className={row()}>
          <div className={field()}>
            <TextField
              size={1}
              type="email"
              value={email}
              disabled={disabled}
              placeholder={t("userForm.fields.additionalEmailPlaceholder")}
              aria-label={t("userForm.fields.additionalEmail", { index: index + 1 })}
              onValueChange={(next) => {
                const updated = [...emails];
                updated[index] = next;
                onValueChange(updated);
              }}
              onBlur={onBlur}
            />
          </div>
          <IconButton
            variant="ghost"
            color="neutral"
            size={1}
            disabled={disabled || (emails.length === 1 && email === "")}
            aria-label={t("userForm.actions.removeEmail")}
            onClick={() => {
              const updated = emails.filter((_, emailIndex) => emailIndex !== index);
              onValueChange(updated.length > 0 ? updated : [""]);
              onBlur?.();
            }}
          >
            <TrashIcon />
          </IconButton>
        </div>
      ))}
      <Button
        type="button"
        variant="ghost"
        color="neutral"
        size={1}
        disabled={disabled}
        iconStart={<PlusIcon />}
        onClick={() => onValueChange([...emails, ""])}
      >
        {t("userForm.actions.addEmail")}
      </Button>
    </div>
  );
}
