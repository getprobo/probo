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

import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import type { NotificationSettingsPageQuery } from "#/__generated__/iam/NotificationSettingsPageQuery.graphql";

import { notificationPreferenceRow } from "../variants";

const NS = "iam/notificationSettings";

export const emailOptions = ["NONE", "EACH", "REMINDERS", "WEEKLY"] as const;

type Preference
  = NotificationSettingsPageQuery["response"]["viewer"]["notificationPreferences"][number];

export type EmailDelivery = Preference["email"];

export interface NotificationPreferenceRowProps {
  preference: Preference;
  onEmailChange: (email: EmailDelivery) => void;
}

export function NotificationPreferenceRow({
  preference,
  onEmailChange,
}: NotificationPreferenceRowProps) {
  const { t } = useTranslation(NS);
  const { control } = notificationPreferenceRow();
  const label = t(`types.${preference.type}.label`);

  return (
    <ListItem>
      <ListItemContent>
        <Text size={2} weight="medium" color="neutral" highContrast>
          {label}
        </Text>
        <Text size={1} color="faint">
          {t(`types.${preference.type}.description`)}
        </Text>
      </ListItemContent>
      <div className={control()}>
        <Select<EmailDelivery>
          value={preference.email}
          onValueChange={(email) => {
            if (email == null) {
              return;
            }
            onEmailChange(email);
          }}
        >
          <SelectTrigger size={2} aria-label={label}>
            {(value: EmailDelivery | null) => (
              value == null ? null : t(`emailOptions.${value}.label`)
            )}
          </SelectTrigger>
          <SelectPopup>
            {emailOptions.map(option => (
              <SelectItem key={option} value={option}>
                {t(`emailOptions.${option}.label`)}
              </SelectItem>
            ))}
          </SelectPopup>
        </Select>
      </div>
    </ListItem>
  );
}
