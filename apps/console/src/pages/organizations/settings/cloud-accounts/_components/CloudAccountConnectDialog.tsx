// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import {
  type CloudAccountProvider,
  getCloudAccountProviderLabel,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Card,
  Dialog,
  DialogContent,
  type DialogRef,
  IconArrowLink,
} from "@probo/ui";
import { type ReactNode, useState } from "react";

import { AwsConnectWizard } from "./AwsConnectWizard";
import { AzureConnectWizard } from "./AzureConnectWizard";
import { GcpConnectWizard } from "./GcpConnectWizard";

type Props = {
  ref: DialogRef;
  connectionId: string;
};

const PROVIDER_BLURB: Record<CloudAccountProvider, string> = {
  AWS: "Connect an AWS account via a Quick-Create CloudFormation stack.",
  GCP: "Connect a GCP project or organization via a setup script and service-account key.",
  AZURE:
    "Connect an Azure subscription via an app registration and client secret.",
};

export function CloudAccountConnectDialog(props: Props) {
  const { ref, connectionId } = props;
  const { __ } = useTranslate();
  const [provider, setProvider] = useState<CloudAccountProvider | null>(null);

  const handleClose = () => {
    setProvider(null);
    ref.current?.close();
  };

  const handleBack = () => {
    setProvider(null);
  };

  const title = provider
    ? `${__("Connect")} — ${getCloudAccountProviderLabel(__, provider)}`
    : __("Connect a cloud account");

  return (
    <Dialog
      ref={ref}
      title={title}
      className="max-w-3xl"
      onClose={() => setProvider(null)}
    >
      <DialogContent padded>
        {provider === null && <ProviderPicker onPick={setProvider} />}
        {provider === "AWS" && (
          <AwsConnectWizard
            connectionId={connectionId}
            onComplete={handleClose}
            onBack={handleBack}
          />
        )}
        {provider === "GCP" && (
          <GcpConnectWizard
            connectionId={connectionId}
            onComplete={handleClose}
            onBack={handleBack}
          />
        )}
        {provider === "AZURE" && (
          <AzureConnectWizard
            connectionId={connectionId}
            onComplete={handleClose}
            onBack={handleBack}
          />
        )}
      </DialogContent>
    </Dialog>
  );
}

function ProviderPicker(props: {
  onPick: (provider: CloudAccountProvider) => void;
}) {
  const { __ } = useTranslate();
  return (
    <div className="space-y-4">
      <p className="text-sm text-txt-secondary">
        {__(
          "Pick a cloud provider to begin. Each wizard walks you through a least-privilege, read-only credential setup.",
        )}
      </p>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
        <ProviderCard
          providerLabel={__("Amazon Web Services")}
          blurb={__(PROVIDER_BLURB.AWS)}
          onClick={() => props.onPick("AWS")}
        />
        <ProviderCard
          providerLabel={__("Google Cloud")}
          blurb={__(PROVIDER_BLURB.GCP)}
          onClick={() => props.onPick("GCP")}
        />
        <ProviderCard
          providerLabel={__("Microsoft Azure")}
          blurb={__(PROVIDER_BLURB.AZURE)}
          onClick={() => props.onPick("AZURE")}
        />
      </div>
    </div>
  );
}

function ProviderCard(props: {
  providerLabel: string;
  blurb: string;
  onClick: () => void;
  icon?: ReactNode;
}) {
  return (
    <Card
      padded
      className="cursor-pointer hover:border-border-mid transition flex flex-col gap-2"
    >
      <button
        type="button"
        onClick={props.onClick}
        className="text-left flex flex-col gap-2 w-full h-full"
      >
        <div className="flex items-center justify-between">
          <h3 className="font-medium">{props.providerLabel}</h3>
          <IconArrowLink size={16} />
        </div>
        <p className="text-xs text-txt-tertiary">{props.blurb}</p>
      </button>
    </Card>
  );
}
