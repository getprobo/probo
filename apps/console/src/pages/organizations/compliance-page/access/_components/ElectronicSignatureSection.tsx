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
    certificateFileUrl
    events {
      id
      eventType
      actorEmail
      occurredAt
    }
  }
`;

function getFilenameFromUrl(url: string): string | null {
  try {
    const disposition = new URL(url).searchParams.get("response-content-disposition");
    const match = disposition?.match(/filename="([^"]+)"/);
    return match ? decodeURIComponent(match[1]) : null;
  } catch {
    return null;
  }
}

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
        {signature.certificateFileUrl && (
          <div className="flex items-center justify-between">
            <span className="text-sm text-txt-secondary">{__("Certificate")}</span>
            <a
              href={signature.certificateFileUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm text-txt-primary hover:underline"
              download
            >
              {getFilenameFromUrl(signature.certificateFileUrl) ?? __("Download")}
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
