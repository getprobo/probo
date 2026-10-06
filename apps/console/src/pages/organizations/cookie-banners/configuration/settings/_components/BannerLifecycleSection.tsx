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

import { TrashIcon } from "@phosphor-icons/react";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { BannerLifecycleSection_cookieBanner$key } from "#/__generated__/core/BannerLifecycleSection_cookieBanner.graphql";
import type { BannerLifecycleSectionActivateMutation } from "#/__generated__/core/BannerLifecycleSectionActivateMutation.graphql";
import type { BannerLifecycleSectionDeactivateMutation } from "#/__generated__/core/BannerLifecycleSectionDeactivateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { cookieBannerLifecycleSection } from "../../../variants";
import { DeleteCookieBannerDialog } from "../../_components/DeleteCookieBannerDialog";

const bannerLifecycleSectionFragment = graphql`
  fragment BannerLifecycleSection_cookieBanner on CookieBanner {
    id
    name
    state
    canDelete: permission(action: "core:cookie-banner:delete")
  }
`;

const activateMutation = graphql`
  mutation BannerLifecycleSectionActivateMutation($input: ActivateCookieBannerInput!) {
    activateCookieBanner(input: $input) {
      cookieBanner {
        id
        state
      }
    }
  }
`;

const deactivateMutation = graphql`
  mutation BannerLifecycleSectionDeactivateMutation($input: DeactivateCookieBannerInput!) {
    deactivateCookieBanner(input: $input) {
      cookieBanner {
        id
        state
      }
    }
  }
`;

interface BannerLifecycleSectionProps {
  cookieBannerKey: BannerLifecycleSection_cookieBanner$key;
}

export function BannerLifecycleSection({
  cookieBannerKey,
}: BannerLifecycleSectionProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const banner = useFragment(bannerLifecycleSectionFragment, cookieBannerKey);
  const { root, intro, row, status, actions } = cookieBannerLifecycleSection();
  const [deleteOpen, setDeleteOpen] = useState(false);

  const [activate, isActivating] = useMutation<BannerLifecycleSectionActivateMutation>(
    activateMutation,
    {
      successMessage: t("configLayout.messages.activated"),
      errorToast: t("configLayout.errors.activate"),
    },
  );
  const [deactivate, isDeactivating] = useMutation<BannerLifecycleSectionDeactivateMutation>(
    deactivateMutation,
    {
      successMessage: t("configLayout.messages.deactivated"),
      errorToast: t("configLayout.errors.deactivate"),
    },
  );

  const isActive = banner.state === "ACTIVE";

  return (
    <section className={root()}>
      <div className={intro()}>
        <Heading level={2} size={4} weight="medium" highContrast>
          {t("bannerLifecycleSection.title")}
        </Heading>
        <Text size={2} color="faint">
          {t("bannerLifecycleSection.description")}
        </Text>
      </div>
      <Card size={2} variant="soft">
        <div className={row()}>
          <div className={status()}>
            <Badge color={isActive ? "green" : "red"} variant="soft">
              {isActive
                ? t("configLayout.status.active")
                : t("configLayout.status.inactive")}
            </Badge>
          </div>
          <div className={actions()}>
            <Button
              size={2}
              variant="soft"
              color="neutral"
              disabled={isActivating || isDeactivating}
              onClick={() => {
                const mutate = isActive ? deactivate : activate;
                void mutate({
                  variables: { input: { cookieBannerId: banner.id } },
                }).catch(() => {
                  // Error toast is already shown by useMutation.
                });
              }}
            >
              {isActive
                ? t("configLayout.actions.deactivate")
                : t("configLayout.actions.activate")}
            </Button>
            {banner.canDelete && !isActive && (
              <Button
                size={2}
                variant="soft"
                color="red"
                iconStart={<TrashIcon />}
                onClick={() => setDeleteOpen(true)}
              >
                {t("configLayout.actions.delete")}
              </Button>
            )}
          </div>
        </div>
      </Card>
      <DeleteCookieBannerDialog
        cookieBannerId={banner.id}
        name={banner.name}
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
      />
    </section>
  );
}
