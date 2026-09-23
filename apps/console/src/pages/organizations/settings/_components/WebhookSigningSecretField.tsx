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

import { CopyIcon } from "@phosphor-icons/react";
import { useToast } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { fetchQuery, graphql, useRelayEnvironment } from "react-relay";

import type { WebhookSigningSecretFieldQuery } from "#/__generated__/core/WebhookSigningSecretFieldQuery.graphql";

import { webhookSigningSecretField } from "../variants";

const webhookSigningSecretFieldQuery = graphql`
  query WebhookSigningSecretFieldQuery($webhookSubscriptionId: ID!) {
    node(id: $webhookSubscriptionId) {
      __typename
      ... on WebhookSubscription {
        signingSecret
      }
    }
  }
`;

const MASKED_SECRET = "••••••••••••••••";

interface WebhookSigningSecretFieldProps {
  webhookSubscriptionId: string;
}

export function WebhookSigningSecretField({
  webhookSubscriptionId,
}: WebhookSigningSecretFieldProps) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const environment = useRelayEnvironment();
  const { root, header, value, actions } = webhookSigningSecretField();
  const [secret, setSecret] = useState<string | null>(null);
  const [revealed, setRevealed] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  async function loadSecret() {
    if (secret != null) {
      return secret;
    }

    setIsLoading(true);
    try {
      const data = await fetchQuery<WebhookSigningSecretFieldQuery>(
        environment,
        webhookSigningSecretFieldQuery,
        { webhookSubscriptionId },
      ).toPromise();
      const next = data?.node?.__typename === "WebhookSubscription"
        ? data.node.signingSecret
        : null;
      if (next == null || next === "") {
        throw new Error("missing signing secret");
      }
      setSecret(next);
      return next;
    } catch {
      toast({
        title: t("webhooksSettingsPage.errorTitle"),
        description: t("webhooksSettingsPage.errors.loadSigningSecret"),
        variant: "error",
      });
      return null;
    } finally {
      setIsLoading(false);
    }
  }

  async function handleToggleReveal() {
    if (revealed) {
      setRevealed(false);
      return;
    }

    const next = await loadSecret();
    if (next != null) {
      setRevealed(true);
    }
  }

  async function handleCopy() {
    const next = await loadSecret();
    if (next == null) {
      return;
    }

    try {
      await navigator.clipboard.writeText(next);
      toast({
        title: t("webhooksSettingsPage.copiedToClipboard"),
        description: t("webhooksSettingsPage.fields.signingSecret"),
        variant: "success",
      });
    } catch {
      toast({
        title: t("webhooksSettingsPage.errorTitle"),
        description: t("webhooksSettingsPage.errors.loadSigningSecret"),
        variant: "error",
      });
    }
  }

  return (
    <div className={root()}>
      <div className={header()}>
        <Text size={2} weight="medium">
          {t("webhooksSettingsPage.fields.signingSecret")}
        </Text>
        <div className={actions()}>
          <Button
            variant="ghost"
            color="neutral"
            size={1}
            disabled={isLoading}
            onClick={() => {
              void handleToggleReveal();
            }}
          >
            {revealed
              ? t("webhooksSettingsPage.actions.hide")
              : t("webhooksSettingsPage.actions.show")}
          </Button>
          <IconButton
            variant="ghost"
            color="neutral"
            size={1}
            loading={isLoading}
            aria-label={t("webhooksSettingsPage.copySigningSecret")}
            onClick={() => {
              void handleCopy();
            }}
          >
            <CopyIcon />
          </IconButton>
        </div>
      </div>
      <Code variant="ghost" className={value()}>
        {revealed && secret != null ? secret : MASKED_SECRET}
      </Code>
    </div>
  );
}
