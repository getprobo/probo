import { useTranslate } from "@probo/i18n";
import { IconCircleCheck, IconRadioUnchecked } from "@probo/ui";
import { clsx } from "clsx";
import { graphql, useFragment } from "react-relay";

import type { VersionRowFragment$key } from "#/__generated__/core/VersionRowFragment.graphql";

const fragment = graphql`
  fragment VersionRowFragment on DocumentVersion {
    id
    version
    signed
    publishedAt
  }
`;

export function VersionRow({
  fKey,
  isSelected,
  onSelect,
}: {
  fKey: VersionRowFragment$key;
  isSelected: boolean;
  onSelect: () => void;
}) {
  const { __ } = useTranslate();
  const versionData = useFragment<VersionRowFragment$key>(fragment, fKey);
  const isVersionSigned = versionData.signed;

  return (
    <div
      onClick={onSelect}
      className={clsx(
        "flex items-center gap-3 py-3 px-4 transition-colors cursor-pointer",
        isSelected
          ? "bg-blue-50 border-l-4 border-blue-500"
          : "bg-transparent hover:bg-level-1",
      )}
    >
      <div className="flex items-center justify-center w-8 h-8 rounded-full bg-level-2 flex-shrink-0">
        {isVersionSigned
          ? (
              <IconCircleCheck size={20} className="text-txt-success" />
            )
          : (
              <IconRadioUnchecked size={20} className="text-txt-tertiary" />
            )}
      </div>
      <div className="flex-1 min-w-0">
        <p
          className={clsx(
            "text-sm font-medium truncate",
            isVersionSigned ? "text-txt-tertiary" : "text-txt-primary",
          )}
        >
          {versionData.publishedAt
            ? `v${versionData.version} - ${(() => {
              const date = new Date(versionData.publishedAt);
              const day = String(date.getDate()).padStart(2, "0");
              const month = String(date.getMonth() + 1).padStart(2, "0");
              const year = date.getFullYear();
              return `${day}/${month}/${year}`;
            })()}`
            : `v${versionData.version}`}
        </p>
      </div>
      <div className="flex-shrink-0">
        <span
          className={clsx(
            "inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium",
            isVersionSigned
              ? "bg-green-100 text-green-800"
              : isSelected
                ? "bg-blue-100 text-blue-800"
                : "bg-gray-100 text-gray-700",
          )}
        >
          {isVersionSigned
            ? __("Signed")
            : isSelected
              ? __("In review")
              : __("Waiting signature")}
        </span>
      </div>
    </div>
  );
}
