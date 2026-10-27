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

import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { NotificationSettingsPageQuery } from "#/__generated__/iam/NotificationSettingsPageQuery.graphql";
import type { NotificationSettingsPageUpdateMutation } from "#/__generated__/iam/NotificationSettingsPageUpdateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";
import { PageHeader } from "#/pages/_components/PageHeader";

import {
  NotificationPreferenceList,
} from "./_components/NotificationPreferenceList";
import type { EmailDelivery } from "./_components/NotificationPreferenceRow";

const NS = "iam/notificationSettings";

export const notificationSettingsPageQuery = graphql`
  query NotificationSettingsPageQuery($organizationId: ID!) {
    viewer @required(action: THROW) {
      notificationPreferences(organizationId: $organizationId) {
        type
        email
      }
    }
  }
`;

const updateNotificationPreferenceMutation = graphql`
  mutation NotificationSettingsPageUpdateMutation($input: UpdateNotificationPreferenceInput!) {
    updateNotificationPreference(input: $input) {
      preference {
        type
        email
      }
    }
  }
`;

type Preference
  = NotificationSettingsPageQuery["response"]["viewer"]["notificationPreferences"][number];

interface NotificationSettingsPageProps {
  organizationId: string;
  queryRef: PreloadedQuery<NotificationSettingsPageQuery>;
}

export function NotificationSettingsPage({
  organizationId,
  queryRef,
}: NotificationSettingsPageProps) {
  const { t } = useTranslation(NS);
  const { t: tApp } = useTranslation();
  const data = usePreloadedQuery<NotificationSettingsPageQuery>(
    notificationSettingsPageQuery,
    queryRef,
  );
  const [preferences, setPreferences] = useState(data.viewer.notificationPreferences);
  const [updatePreference] = useMutation<NotificationSettingsPageUpdateMutation>(
    updateNotificationPreferenceMutation,
  );

  function save(preference: Preference, email: EmailDelivery) {
    const next = { ...preference, email };
    const previous = preferences;
    setPreferences(current => current.map(item => (
      item.type === next.type ? next : item
    )));

    void updatePreference({
      variables: {
        input: {
          organizationId,
          type: next.type,
          email: next.email,
        },
      },
    }).catch(() => {
      setPreferences(previous);
    });
  }

  return (
    <main className="mx-auto flex w-full max-w-5xl flex-col gap-10 px-8 pt-8 pb-32">
      <PageHeader
        homeLabel={tApp("homePage.breadcrumb")}
        currentLabel={t("breadcrumb")}
        title={t("title")}
      />
      <NotificationPreferenceList
        preferences={preferences}
        onEmailChange={save}
      />
    </main>
  );
}
