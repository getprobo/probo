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
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import type { PublishCookieBannerDialogPublishMutation } from "#/__generated__/core/PublishCookieBannerDialogPublishMutation.graphql";
import type { PublishCookieBannerDialogRegeneratePolicyMutation } from "#/__generated__/core/PublishCookieBannerDialogRegeneratePolicyMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const publishMutation = graphql`
  mutation PublishCookieBannerDialogPublishMutation($input: PublishCookieBannerVersionInput!) {
    publishCookieBannerVersion(input: $input) {
      cookieBannerVersion {
        id
        version
        state
      }
      cookieBanner {
        ...CookieBannerConfigLayout_cookieBanner
        publishedVersion {
          id
          gvlVendorCount
          gvlVendorIds
        }
      }
    }
  }
`;

const regeneratePolicyMutation = graphql`
  mutation PublishCookieBannerDialogRegeneratePolicyMutation(
    $input: RegenerateCookieBannerTrackerPolicyInput!
  ) {
    regenerateCookieBannerTrackerPolicy(input: $input) {
      cookieBanner {
        id
        ...CookieBannerConfigLayout_cookieBanner
      }
    }
  }
`;

interface PublishCookieBannerDialogProps {
  cookieBannerId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function PublishCookieBannerDialog({
  cookieBannerId,
  open,
  onOpenChange,
}: PublishCookieBannerDialogProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const [pending, setPending] = useState<"only" | "policy" | null>(null);
  const [publish] = useMutation<PublishCookieBannerDialogPublishMutation>(
    publishMutation,
    {
      successMessage: t("configLayout.messages.published"),
      errorToast: t("configLayout.errors.publish"),
    },
  );
  const [regeneratePolicy] = useMutation<PublishCookieBannerDialogRegeneratePolicyMutation>(
    regeneratePolicyMutation,
    {
      errorToast: t("configLayout.errors.regeneratePolicy"),
    },
  );
  const busy = pending != null;

  function handleOpenChange(next: boolean) {
    if (busy) return;
    onOpenChange(next);
  }

  async function handlePublish(updatePolicy: boolean) {
    setPending(updatePolicy ? "policy" : "only");
    let published = false;
    try {
      await publish({
        variables: { input: { cookieBannerId } },
      });
      published = true;
      if (updatePolicy) {
        await regeneratePolicy(
          { variables: { input: { cookieBannerId } } },
          { successMessage: t("configLayout.messages.policyRequested") },
        );
      }
      onOpenChange(false);
    } catch {
      if (published) {
        onOpenChange(false);
      }
    } finally {
      setPending(null);
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("configLayout.dialogs.publishPolicyTitle")}</DialogTitle>
          <DialogDescription>
            {t("configLayout.dialogs.publishPolicyDescription")}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral" disabled={busy}>
                {t("configLayout.actions.cancel")}
              </Button>
            )}
          />
          <Button
            variant="soft"
            color="neutral"
            disabled={busy}
            loading={pending === "only"}
            onClick={() => {
              void handlePublish(false);
            }}
          >
            {t("configLayout.actions.publishOnly")}
          </Button>
          <Button
            variant="solid"
            color="neutral"
            highContrast
            disabled={busy}
            loading={pending === "policy"}
            onClick={() => {
              void handlePublish(true);
            }}
          >
            {t("configLayout.actions.publishAndUpdatePolicy")}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
