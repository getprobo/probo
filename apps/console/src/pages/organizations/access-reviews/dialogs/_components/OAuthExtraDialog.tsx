// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Option,
  Select,
  useDialogRef,
} from "@probo/ui";
import { useEffect, useState } from "react";

import {
  cleanZendeskSubdomain,
  connectOAuthProvider,
  DATADOG_SITES,
} from "../_lib/connectorSettings";
import type { ProviderInfo } from "../AddAccessReviewSourceDialog";

type OAuthExtraDialogProps = {
  provider: ProviderInfo | null;
  organizationId: string;
  onClose: () => void;
};

export function DatadogConnectDialog({
  provider,
  organizationId,
  onClose,
}: OAuthExtraDialogProps) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const [datadogSite, setDatadogSite] = useState<string>("US1");

  // Opening is driven imperatively by the parent's active-provider state; the
  // form is reset on close so opening only shows the dialog (no setState here).
  useEffect(() => {
    if (provider) {
      dialogRef.current?.open();
    }
  }, [provider]);

  return (
    <Dialog
      ref={dialogRef}
      onClose={() => {
        setDatadogSite("US1");
        onClose();
      }}
      title={__("Connect Datadog")}
    >
      <form
        onSubmit={(e) => {
          e.preventDefault();
          if (provider) {
            connectOAuthProvider(organizationId, provider, { site: datadogSite });
          }
        }}
      >
        <DialogContent padded className="space-y-4">
          <p className="text-txt-secondary text-sm">
            {__("Select your Datadog site, then continue to authorize access.")}
          </p>
          <div className="space-y-1.5">
            <label className="text-sm font-medium">{__("Datadog site")}</label>
            <Select
              value={datadogSite}
              onValueChange={setDatadogSite}
              placeholder={__("Select a site")}
            >
              {DATADOG_SITES.map(s => (
                <Option key={s.value} value={s.value}>
                  {s.label}
                </Option>
              ))}
            </Select>
          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit">{__("Continue")}</Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}

export function ZendeskConnectDialog({
  provider,
  organizationId,
  onClose,
}: OAuthExtraDialogProps) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const [zendeskSubdomain, setZendeskSubdomain] = useState<string>("");

  // Opening is driven imperatively by the parent's active-provider state; the
  // form is reset on close so opening only shows the dialog (no setState here).
  useEffect(() => {
    if (provider) {
      dialogRef.current?.open();
    }
  }, [provider]);

  return (
    <Dialog
      ref={dialogRef}
      onClose={() => {
        setZendeskSubdomain("");
        onClose();
      }}
      title={__("Connect Zendesk")}
    >
      <form
        onSubmit={(e) => {
          e.preventDefault();
          if (provider) {
            const site = cleanZendeskSubdomain(zendeskSubdomain);
            if (site) {
              connectOAuthProvider(organizationId, provider, { site });
            }
          }
        }}
      >
        <DialogContent padded className="space-y-4">
          <p className="text-txt-secondary text-sm">
            {__("Enter your Zendesk subdomain, then continue to authorize access.")}
          </p>
          <Field
            label={__("Zendesk subdomain")}
            placeholder={__("acme")}
            value={zendeskSubdomain}
            onValueChange={setZendeskSubdomain}
            help={__("The <subdomain> part of <subdomain>.zendesk.com")}
            required
            autoFocus
          />
        </DialogContent>
        <DialogFooter>
          <Button
            type="submit"
            disabled={!cleanZendeskSubdomain(zendeskSubdomain)}
          >
            {__("Continue")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
