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
import { useQueryLoader } from "react-relay";
import { useNavigate, useParams } from "react-router";

import type { ThirdPartySwitcherMenuQuery } from "#/__generated__/core/ThirdPartySwitcherMenuQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import {
  navPanelSwitcher,
  NavPanelSwitcher,
  NavPanelSwitcherValue,
  NavPanelSwitcherValueSkeleton,
} from "#/pages/organizations/_components/NavPanelSwitcher";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { thirdPartyHref } from "../_lib/thirdPartySections";

import { CreateThirdPartyDialog } from "./CreateThirdPartyDialog";
import { ThirdPartyNavItems } from "./ThirdPartyNavItems";
import {
  ThirdPartySwitcherMenu,
  thirdPartySwitcherMenuQuery,
} from "./ThirdPartySwitcherMenu";
import { ThirdPartySwitcherValue } from "./ThirdPartySwitcherValue";

export function ThirdPartySwitcher() {
  return (
    <CoreRelayProvider>
      <ThirdPartySwitcherInner />
      <ThirdPartyNavItems />
    </CoreRelayProvider>
  );
}

function ThirdPartySwitcherInner() {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const { thirdPartyId } = useParams<{ thirdPartyId: string }>();
  const [queryRef, loadQuery] = useQueryLoader<ThirdPartySwitcherMenuQuery>(
    thirdPartySwitcherMenuQuery,
  );
  const [createConnection, setCreateConnection] = useState<string | null>(null);

  const selectLabel = t("nav.thirdPartySwitcher.select");
  const slots = navPanelSwitcher();

  const handleOpenChange = useCallback((open: boolean) => {
    if (open) {
      loadQuery({ organizationId });
    }
  }, [loadQuery, organizationId]);

  const handleCreated = useCallback((id: string) => {
    void navigate(thirdPartyHref(organizationId, id));
  }, [navigate, organizationId]);

  return (
    <>
      <NavPanelSwitcher
        active={false}
        onOpenChange={handleOpenChange}
        value={thirdPartyId != null
          ? (
              <Suspense
                fallback={(
                  <>
                    <AvatarSkeleton size={1} radius="small" />
                    <NavPanelSwitcherValueSkeleton />
                  </>
                )}
              >
                <ThirdPartySwitcherValue fallback={selectLabel} />
              </Suspense>
            )
          : (
              <NavPanelSwitcherValue>
                {selectLabel}
              </NavPanelSwitcherValue>
            )}
      >
        {queryRef != null && (
          <Suspense
            fallback={(
              <Text size={2} color="faint" className={slots.empty()}>
                {t("nav.thirdPartySwitcher.loading")}
              </Text>
            )}
          >
            <ThirdPartySwitcherMenu
              queryRef={queryRef}
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
