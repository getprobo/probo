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

import { AvatarSkeleton } from "@probo/ui/src/v2/Avatar/AvatarSkeleton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense, useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, useQueryLoader } from "react-relay";
import { useNavigate, useParams } from "react-router";

import type { ThirdPartySwitcherMenuQuery } from "#/__generated__/core/ThirdPartySwitcherMenuQuery.graphql";
import type { ThirdPartySwitcherValueQuery } from "#/__generated__/core/ThirdPartySwitcherValueQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import {
  navPanelSwitcher,
  NavPanelSwitcher,
  NavPanelSwitcherValue,
  NavPanelSwitcherValueSkeleton,
} from "#/pages/organizations/_components/NavPanelSwitcher";

import { thirdPartyHref } from "../_lib/thirdPartySections";

import { CreateThirdPartyDialog } from "./CreateThirdPartyDialog";
import {
  ThirdPartySwitcherMenu,
  thirdPartySwitcherMenuQuery,
} from "./ThirdPartySwitcherMenu";
import { ThirdPartySwitcherValue } from "./ThirdPartySwitcherValue";

export interface ThirdPartySwitcherProps {
  queryRef: PreloadedQuery<ThirdPartySwitcherValueQuery> | null;
}

export function ThirdPartySwitcher({ queryRef }: ThirdPartySwitcherProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const { thirdPartyId } = useParams<{ thirdPartyId: string }>();
  const [menuQueryRef, loadMenuQuery] = useQueryLoader<ThirdPartySwitcherMenuQuery>(
    thirdPartySwitcherMenuQuery,
  );
  const [createConnection, setCreateConnection] = useState<string | null>(null);

  const selectLabel = t("nav.thirdPartySwitcher.select");
  const slots = navPanelSwitcher();

  const handleOpenChange = useCallback((open: boolean) => {
    if (open) {
      loadMenuQuery({ organizationId }, { fetchPolicy: "network-only" });
    }
  }, [loadMenuQuery, organizationId]);

  const handleCreated = useCallback((id: string) => {
    void navigate(thirdPartyHref(organizationId, id));
  }, [navigate, organizationId]);

  // The parent loads the query in an effect, so on a direct third-party to
  // third-party navigation the ref still points at the previous one.
  const currentQueryRef = queryRef?.variables.thirdPartyId === thirdPartyId
    ? queryRef
    : null;

  const value = thirdPartyId != null && currentQueryRef != null
    ? (
        <Suspense
          fallback={(
            <>
              <AvatarSkeleton size={1} radius="small" />
              <NavPanelSwitcherValueSkeleton />
            </>
          )}
        >
          <ThirdPartySwitcherValue fallback={selectLabel} queryRef={currentQueryRef} />
        </Suspense>
      )
    : thirdPartyId != null
      ? (
          <>
            <AvatarSkeleton size={1} radius="small" />
            <NavPanelSwitcherValueSkeleton />
          </>
        )
      : (
          <NavPanelSwitcherValue>
            {selectLabel}
          </NavPanelSwitcherValue>
        );

  return (
    <>
      <NavPanelSwitcher
        active={false}
        label={t("nav.thirdPartySwitcher.label")}
        onOpenChange={handleOpenChange}
        value={value}
      >
        {menuQueryRef != null && (
          <Suspense
            fallback={(
              <Text size={2} color="faint" className={slots.empty()}>
                {t("nav.thirdPartySwitcher.loading")}
              </Text>
            )}
          >
            <ThirdPartySwitcherMenu
              queryRef={menuQueryRef}
              onCreate={setCreateConnection}
            />
          </Suspense>
        )}
      </NavPanelSwitcher>
      {createConnection != null && (
        <CreateThirdPartyDialog
          organizationId={organizationId}
          connection={createConnection}
          defaultOpen
          onOpenChange={(open) => {
            if (!open) {
              setCreateConnection(null);
            }
          }}
          onCreated={handleCreated}
        />
      )}
    </>
  );
}
