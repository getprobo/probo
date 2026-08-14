// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { useCopy } from "@probo/hooks";
import { Button, DialogContent } from "@probo/ui";
import { useTranslation } from "react-i18next";

export function SAMLDomainVerifyDialog(props: {
  domainVerificationToken: string;
}) {
  const { domainVerificationToken } = props;

  const { t } = useTranslation();

  const dnsRecord = `probo-verification=${domainVerificationToken}`;
  const [isCopied, copy] = useCopy();

  return (
    <>
      <DialogContent padded className="space-y-6">
        <div>
          <h3 className="text-base font-medium mb-4">
            {t("samlDomainVerifyDialog.title")}
          </h3>
          <p className="text-sm text-gray-600 mb-4">
            {t("samlDomainVerifyDialog.description")}
          </p>
          <div className="bg-gray-50 rounded-lg p-4 mb-4">
            <div className="space-y-2">
              <div>
                <span className="font-semibold text-sm">
                  {t("samlDomainVerifyDialog.fields.hostName")}
                </span>
                <code className="ml-2 bg-white px-2 py-1 rounded text-sm">
                  @
                </code>
                <span className="ml-2 text-xs text-gray-600">
                  {t("samlDomainVerifyDialog.hostHint")}
                </span>
              </div>
              <div>
                <span className="font-semibold text-sm">
                  {t("samlDomainVerifyDialog.fields.type")}
                </span>
                <code className="ml-2 bg-white px-2 py-1 rounded text-sm">
                  TXT
                </code>
              </div>
              <div>
                <span className="font-semibold text-sm">
                  {t("samlDomainVerifyDialog.fields.value")}
                </span>
                <div className="mt-1 flex items-center gap-2">
                  <code className="flex-1 bg-white px-2 py-1 rounded text-sm break-all font-mono">
                    {dnsRecord}
                  </code>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => copy(dnsRecord)}
                  >
                    {isCopied
                      ? t("samlDomainVerifyDialog.actions.copied")
                      : t("samlDomainVerifyDialog.actions.copy")}
                  </Button>
                </div>
              </div>
            </div>
          </div>
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <p className="text-sm text-blue-800">
              <strong>{t("samlDomainVerifyDialog.note.label")}</strong>
              {" "}
              {t("samlDomainVerifyDialog.note.description")}
            </p>
          </div>
        </div>
      </DialogContent>
    </>
  );
}
