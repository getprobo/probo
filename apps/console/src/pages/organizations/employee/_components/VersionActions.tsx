import { useTranslate } from "@probo/i18n";
import { Button, Spinner } from "@probo/ui";
import { graphql, useFragment } from "react-relay";

import type { VersionActionsFragment$key } from "#/__generated__/core/VersionActionsFragment.graphql";

const fragment = graphql`
  fragment VersionActionsFragment on DocumentVersion {
    id
    signed
  }
`;

export function VersionActions({
  fKey,
  isSigning,
  onSign,
  onBack,
}: {
  fKey: VersionActionsFragment$key;
  isSigning: boolean;
  onSign: (versionId: string) => void;
  onBack: () => void;
}) {
  const { __ } = useTranslate();
  const versionData = useFragment<VersionActionsFragment$key>(fragment, fKey);
  const isSigned = versionData.signed;

  if (isSigned) {
    return (
      <>
        <Button onClick={onBack} className="h-10 w-full" variant="secondary">
          {__("Back to Documents")}
        </Button>
        <p className="text-xs text-txt-tertiary mt-2 h-5" />
      </>
    );
  }

  return (
    <>
      <Button
        onClick={() => onSign(versionData.id)}
        className="h-10 w-full"
        disabled={isSigning}
        icon={isSigning ? Spinner : undefined}
      >
        {__("I acknowledge and agree")}
      </Button>
      <p className="text-xs text-txt-tertiary mt-2 h-5">
        {__(
          "By clicking 'I acknowledge and agree', your digital signature will be recorded.",
        )}
      </p>
    </>
  );
}
