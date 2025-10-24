import { graphql } from "relay-runtime";
import type { TrustCenterFileRowFragment$key } from "./__generated__/TrustCenterFileRowFragment.graphql";
import { useFragment } from "react-relay";
import {
  Button,
  IconArrowInbox,
  IconLock,
  IconPageTextLine,
  Spinner,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { TrustCenterFileRowDownloadMutation } from "./__generated__/TrustCenterFileRowDownloadMutation.graphql";
import { useMutationWithToasts } from "/hooks/useMutationWithToast";
import { downloadFile } from "@probo/helpers";
import { RequestAccessDialog } from "/components/RequestAccessDialog.tsx";
import { useState } from "react";

const downloadMutation = graphql`
  mutation TrustCenterFileRowDownloadMutation($input: ExportTrustCenterFileInput!) {
    exportTrustCenterFile(input: $input) {
      data
    }
  }
`;

const trustCenterFileRowFragment = graphql`
  fragment TrustCenterFileRowFragment on TrustCenterFile {
    id
    name
    isUserAuthorized
    hasUserRequestedAccess
  }
`;

export function TrustCenterFileRow(props: { file: TrustCenterFileRowFragment$key }) {
  const file = useFragment(trustCenterFileRowFragment, props.file);
  const { __ } = useTranslate();
  const [commitDownload, downloading] =
    useMutationWithToasts<TrustCenterFileRowDownloadMutation>(downloadMutation);
  const handleDownload = () => {
    commitDownload({
      variables: {
        input: {
          trustCenterFileId: file.id,
        },
      },
      onSuccess(response) {
        downloadFile(response.exportTrustCenterFile.data, file.name);
      },
    });
  };
  const [hasRequested, setHasRequested] = useState(
    file.hasUserRequestedAccess,
  );
  return (
    <div className="text-sm border-1 border-border-solid -mt-[1px] flex gap-3 flex-col md:flex-row md:justify-between px-6 py-3">
      <div className="flex items-center gap-2">
        <IconPageTextLine size={16} className=" flex-none text-txt-tertiary" />
        {file.name}
      </div>
      {file.isUserAuthorized ? (
        <Button
          className="w-full md:w-max"
          variant="secondary"
          disabled={downloading}
          icon={downloading ? Spinner : IconArrowInbox}
          onClick={handleDownload}
        >
          {__("Download")}
        </Button>
      ) : (
        <RequestAccessDialog
          trustCenterFileId={file.id}
          onSuccess={() => setHasRequested(true)}
        >
          <Button
            disabled={hasRequested}
            className="w-full md:w-max"
            variant="secondary"
            icon={IconLock}
          >
            {hasRequested ? __("Access requested") : __("Request access")}
          </Button>
        </RequestAccessDialog>
      )}
    </div>
  );
}
