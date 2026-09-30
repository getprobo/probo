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

import { ArrowsClockwiseIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ReactivateSCIMBridgeButtonFragment$key } from "#/__generated__/iam/ReactivateSCIMBridgeButtonFragment.graphql";
import type { ReactivateSCIMBridgeButtonMutation } from "#/__generated__/iam/ReactivateSCIMBridgeButtonMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { scimProviderCard } from "../variants";

const reactivateSCIMBridgeButtonFragment = graphql`
  fragment ReactivateSCIMBridgeButtonFragment on SCIMBridge {
    id
    state
    canUpdate: permission(action: "iam:scim-bridge:update")
  }
`;

const reactivateSCIMBridgeMutation = graphql`
  mutation ReactivateSCIMBridgeButtonMutation($input: ReactivateSCIMBridgeInput!) {
    reactivateSCIMBridge(input: $input) {
      scimBridge {
        id
        state
        syncError
      }
    }
  }
`;

export interface ReactivateSCIMBridgeButtonProps {
  bridgeKey: ReactivateSCIMBridgeButtonFragment$key;
}

export function ReactivateSCIMBridgeButton({
  bridgeKey,
}: ReactivateSCIMBridgeButtonProps) {
  const bridge = useFragment(reactivateSCIMBridgeButtonFragment, bridgeKey);
  const { t } = useTranslation();
  const { reactivate } = scimProviderCard();
  const [reactivateSCIMBridge, isReactivating]
    = useMutation<ReactivateSCIMBridgeButtonMutation>(
      reactivateSCIMBridgeMutation,
      {
        successMessage: t("reactivateSCIMBridge.messages.success"),
        errorToast: t("reactivateSCIMBridge.errors.reactivate"),
      },
    );

  if (!bridge.canUpdate || bridge.state === "ACTIVE") {
    return null;
  }

  function handleReactivate() {
    void reactivateSCIMBridge({
      variables: {
        input: {
          scimBridgeId: bridge.id,
        },
      },
    }).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  return (
    <div className={reactivate()}>
      <Text size={2} weight="medium" highContrast>
        {t("reactivateSCIMBridge.dialog.title")}
      </Text>
      <Text size={2} color="faint">
        {t("reactivateSCIMBridge.dialog.description")}
      </Text>
      <Button
        variant="soft"
        color="neutral"
        loading={isReactivating}
        iconStart={<ArrowsClockwiseIcon />}
        onClick={handleReactivate}
      >
        {t("reactivateSCIMBridge.actions.reactivate")}
      </Button>
    </div>
  );
}
