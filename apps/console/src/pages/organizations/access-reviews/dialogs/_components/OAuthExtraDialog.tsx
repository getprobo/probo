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
import { useTranslation } from "react-i18next";

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
  const { t } = useTranslation();
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
      title={t("oauthExtraDialog.datadog.title")}
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
            {t("oauthExtraDialog.datadog.description")}
          </p>
          <div className="space-y-1.5">
            <label className="text-sm font-medium">
              {t("oauthExtraDialog.datadog.siteLabel")}
            </label>
            <Select
              value={datadogSite}
              onValueChange={setDatadogSite}
              placeholder={t("oauthExtraDialog.datadog.sitePlaceholder")}
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
          <Button type="submit">{t("oauthExtraDialog.actions.continue")}</Button>
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
  const { t } = useTranslation();
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
      title={t("oauthExtraDialog.zendesk.title")}
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
            {t("oauthExtraDialog.zendesk.description")}
          </p>
          <Field
            label={t("oauthExtraDialog.zendesk.subdomainLabel")}
            placeholder={t("oauthExtraDialog.zendesk.subdomainPlaceholder")}
            value={zendeskSubdomain}
            onValueChange={setZendeskSubdomain}
            help={t("oauthExtraDialog.zendesk.subdomainHelp")}
            required
            autoFocus
          />
        </DialogContent>
        <DialogFooter>
          <Button
            type="submit"
            disabled={!cleanZendeskSubdomain(zendeskSubdomain)}
          >
            {t("oauthExtraDialog.actions.continue")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
