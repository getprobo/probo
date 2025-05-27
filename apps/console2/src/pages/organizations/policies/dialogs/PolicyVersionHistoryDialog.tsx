import {
  Avatar,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogTitle,
  Markdown,
} from "@probo/ui";
import { graphql } from "relay-runtime";
import { useTranslate } from "@probo/i18n";
import { useState, type ReactNode } from "react";
import type {
  PolicyPagePolicyFragment$data,
  PolicyPagePolicyFragment$key,
} from "../__generated__/PolicyPagePolicyFragment.graphql";
import { useFragment } from "react-relay";
import type {
  PolicyVersionHistoryDialogFragment$data,
  PolicyVersionHistoryDialogFragment$key,
} from "./__generated__/PolicyVersionHistoryDialogFragment.graphql";
import clsx from "clsx";
import type { NodeOf } from "/types";

const historyFragment = graphql`
  fragment PolicyVersionHistoryDialogFragment on PolicyVersion {
    id
    version
    status
    content
    changelog
    publishedAt
    updatedAt
    publishedBy {
      fullName
    }
  }
`;

type Props = {
  policy: PolicyPagePolicyFragment$data;
  children?: ReactNode;
};

type Version = NodeOf<PolicyPagePolicyFragment$data["versions"]>;

export function PolicyVersionHistoryDialog(props: Props) {
  const { __ } = useTranslate();
  const versions = props.policy.versions.edges.map((edge) => edge.node);
  const [selectedVersion, setSelectedVersion] = useState<Version>(versions[0]);
  return (
    <Dialog defaultOpen trigger={props.children}>
      <DialogContent className="flex" scrollableChildren>
        <aside className="p-6 overflow-y-auto w-60 flex-none">
          <DialogTitle className="text-base text-txt-primary font-medium mb-4">
            {__("Version History")}
          </DialogTitle>
          {versions.map((version) => (
            <VersionItem
              key={version.id}
              version={version}
              active={selectedVersion === version}
              onSelect={setSelectedVersion}
            />
          ))}
        </aside>
        <main className="flex-1 px-12 py-8">
          <Markdown content={selectedVersion.content} />
        </main>
      </DialogContent>
      <DialogFooter></DialogFooter>
    </Dialog>
  );
}

function VersionItem(props: {
  version: Version;
  active?: boolean;
  onSelect: (v: Version) => void;
}) {
  const version = useFragment<PolicyVersionHistoryDialogFragment$key>(
    historyFragment,
    props.version
  );
  const { dateFormat } = useTranslate();
  return (
    <button
      onClick={() => props.onSelect(props.version)}
      className={clsx(
        "flex items-center gap-1 py-2 px-[10px] w-full hover:bg-tertiary-hover cursor-pointer rounded",
        props.active && "bg-tertiary-pressed"
      )}
    >
      <Avatar name={version.publishedBy?.fullName ?? ""} size="l" />
      <div className="space-y-[2px]">
        <div className="text-sm text-txt-primary">
          {version.publishedBy?.fullName ?? "Anonymous"}
        </div>
        <div className="text-xs text-txt-secondary">
          {dateFormat(version.publishedAt, {
            day: "numeric",
            month: "short",
            year: "numeric",
          })}
        </div>
      </div>
    </button>
  );
}
