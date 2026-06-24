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

import { formatDatetime } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ElectronicSignatureSectionFragment$key } from "#/__generated__/core/ElectronicSignatureSectionFragment.graphql";

import { EventTypeLabel } from "./EventTypeLabel";
import { NdaSignatureBadge } from "./NdaSignatureBadge";

const fragment = graphql`
  fragment ElectronicSignatureSectionFragment on ElectronicSignature {
    status
    signedAt
    certificate {
      downloadUrl
      fileName
    }
    events {
      id
      eventType
      actorEmail
      occurredAt
    }
  }
`;

export function ElectronicSignatureSection({
  fragmentRef,
}: {
  fragmentRef: ElectronicSignatureSectionFragment$key;
}) {
  const { __ } = useTranslate();
  const signature = useFragment(fragment, fragmentRef);

  return (
    <div>
      <h3 className="text-sm font-medium text-txt-primary mb-3">
        {__("Electronic Signature")}
      </h3>
      <div className="rounded-lg border border-border-solid bg-bg-secondary p-4 space-y-3">
        <div className="flex items-center justify-between">
          <span className="text-sm text-txt-secondary">{__("Status")}</span>
          <NdaSignatureBadge status={signature.status} />
        </div>
        {signature.signedAt && (
          <div className="flex items-center justify-between">
            <span className="text-sm text-txt-secondary">{__("Signed at")}</span>
            <span className="text-sm text-txt-primary">
              {formatDatetime(signature.signedAt)}
            </span>
          </div>
        )}
        {signature.certificate?.downloadUrl && (
          <div className="flex items-center justify-between">
            <span className="text-sm text-txt-secondary">{__("Certificate")}</span>
            <a
              href={signature.certificate.downloadUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm text-txt-primary hover:underline"
              download
            >
              {signature.certificate.fileName ?? __("Download")}
            </a>
          </div>
        )}
        {signature.events.length > 0 && (
          <div className="pt-2 border-t border-border-solid">
            <span className="text-xs font-medium text-txt-secondary uppercase tracking-wider">
              {__("Activity")}
            </span>
            <div className="mt-2 space-y-2">
              {signature.events.map(event => (
                <div
                  key={event.id}
                  className="flex items-start justify-between text-xs"
                >
                  <div>
                    <span className="text-txt-primary">
                      <EventTypeLabel eventType={event.eventType} />
                    </span>
                    <span className="text-txt-tertiary ml-1">
                      {event.actorEmail}
                    </span>
                  </div>
                  <span className="text-txt-tertiary shrink-0 ml-2">
                    {formatDatetime(event.occurredAt)}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
