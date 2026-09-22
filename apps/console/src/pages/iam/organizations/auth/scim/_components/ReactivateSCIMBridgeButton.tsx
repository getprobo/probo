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

import { Button, IconArrowsClockwise, useToast } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useMutation } from "react-relay";

import type { ReactivateSCIMBridgeButtonFragment$key } from "#/__generated__/iam/ReactivateSCIMBridgeButtonFragment.graphql";
import type { ReactivateSCIMBridgeButtonMutation } from "#/__generated__/iam/ReactivateSCIMBridgeButtonMutation.graphql";

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

export function ReactivateSCIMBridgeButton(props: { fKey: ReactivateSCIMBridgeButtonFragment$key }) {
  const { fKey } = props;
  const bridge = useFragment(reactivateSCIMBridgeButtonFragment, fKey);
  const { t } = useTranslation();
  const { toast } = useToast();

  const [reactivateSCIMBridge, isReactivating]
    = useMutation<ReactivateSCIMBridgeButtonMutation>(
      reactivateSCIMBridgeMutation,
    );

  if (!bridge.canUpdate || bridge.state === "ACTIVE") {
    return null;
  }

  const handleReactivate = () => {
    void reactivateSCIMBridge({
      variables: {
        input: {
          scimBridgeId: bridge.id,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: t("common.error"),
            description: errors.map(e => e.message).join(", "),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("common.success"),
          description: t("reactivateSCIMBridge.messages.success"),
          variant: "success",
        });
      },
      onError(error) {
        toast({
          title: t("common.error"),
          description: error.message,
          variant: "error",
        });
      },
    });
  };

  return (
    <div className="space-y-2 border-t border-border-low pt-6">
      <h4 className="text-sm font-medium">{t("reactivateSCIMBridge.dialog.title")}</h4>
      <p className="text-sm text-txt-secondary">
        {t("reactivateSCIMBridge.dialog.description")}
      </p>
      <Button variant="secondary" onClick={handleReactivate} disabled={isReactivating}>
        <IconArrowsClockwise size={16} />
        {isReactivating
          ? t("reactivateSCIMBridge.actions.reactivating")
          : t("reactivateSCIMBridge.actions.reactivate")}
      </Button>
    </div>
  );
}
