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

import { List } from "@probo/ui/src/v2/List/List";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import type { NotificationSettingsPageQuery } from "#/__generated__/iam/NotificationSettingsPageQuery.graphql";

import { notificationPreferenceList } from "../variants";

import {
  type EmailDelivery,
  emailOptions,
  NotificationPreferenceRow,
} from "./NotificationPreferenceRow";

const NS = "iam/notificationSettings";

type Preference
  = NotificationSettingsPageQuery["response"]["viewer"]["notificationPreferences"][number];

export interface NotificationPreferenceListProps {
  preferences: ReadonlyArray<Preference>;
  onEmailChange: (preference: Preference, email: EmailDelivery) => void;
}

export function NotificationPreferenceList({
  preferences,
  onEmailChange,
}: NotificationPreferenceListProps) {
  const { t } = useTranslation(NS);
  const { root, heading, body, legend, legendItem } = notificationPreferenceList();

  return (
    <section className={root()}>
      <div className={heading()}>
        <Heading level={2} size={2} weight="medium" highContrast>
          {t("email.heading")}
        </Heading>
        <Text size={2} color="faint">
          {t("email.description")}
        </Text>
      </div>
      <div className={body()}>
        <List>
          {preferences.map(preference => (
            <NotificationPreferenceRow
              key={preference.type}
              preference={preference}
              onEmailChange={email => onEmailChange(preference, email)}
            />
          ))}
        </List>
        <dl className={legend()}>
          {emailOptions.map(option => (
            <div key={option} className={legendItem()}>
              <dt>
                <Text size={1} weight="medium" color="neutral">
                  {t(`emailOptions.${option}.label`)}
                </Text>
              </dt>
              <dd>
                <Text size={1} color="faint">
                  {t(`emailOptions.${option}.description`)}
                </Text>
              </dd>
            </div>
          ))}
        </dl>
      </div>
    </section>
  );
}
