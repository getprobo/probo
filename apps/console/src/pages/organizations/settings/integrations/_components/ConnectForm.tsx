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

import { useToast } from "@probo/ui";
import { type ReactNode, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { useNavigate } from "react-router";

import type { ConnectForm_provider$key } from "#/__generated__/core/ConnectForm_provider.graphql";

import type { ConnectDestination } from "../_lib/connectDestination";
import { useConnectorName } from "../_lib/useConnectorName";

import { ConnectFormFooter } from "./ConnectFormFooter";
import { ConnectorNameField } from "./ConnectorNameField";

const connectFormFragment = graphql`
  fragment ConnectForm_provider on ConnectorProviderInfo {
    ...ConnectFormFooter_provider
  }
`;

interface ConnectFormProps {
  providerKey: ConnectForm_provider$key;
  canSubmit?: boolean;
  children?: ReactNode;
  onSubmit: (values: { name: string }) => Promise<ConnectDestination | null> | ConnectDestination | null;
}

export function ConnectForm({
  providerKey,
  canSubmit = true,
  children,
  onSubmit,
}: ConnectFormProps) {
  const { t } = useTranslation("organizations/settings/integrations");
  const { toast } = useToast();
  const navigate = useNavigate();
  const provider = useFragment(connectFormFragment, providerKey);
  const connectorName = useConnectorName();
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit() {
    if (connectorName.rejectIfEmpty() || !canSubmit) {
      return;
    }

    setSubmitting(true);
    try {
      const result = await onSubmit({ name: connectorName.trimmed });
      if (result == null) {
        return;
      }
      toast({
        title: t("marketplacePage.connected"),
        description: t("listPage.messages.connectedDescription"),
        variant: "success",
      });
      void navigate(result.to, { state: result.state ?? null });
    } catch {
      return;
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form
      className="flex flex-col gap-4"
      onSubmit={(event) => {
        event.preventDefault();
        void handleSubmit();
      }}
    >
      <ConnectorNameField
        name={connectorName.name}
        error={connectorName.error}
        onChange={connectorName.onChange}
        onEmpty={connectorName.rejectIfEmpty}
      />
      {children}
      <ConnectFormFooter
        providerKey={provider}
        disabled={!canSubmit}
        loading={submitting}
      />
    </form>
  );
}
