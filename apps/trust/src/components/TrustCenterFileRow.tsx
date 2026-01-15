import { graphql } from "relay-runtime";
import type { TrustCenterFileRowFragment$key } from "./__generated__/TrustCenterFileRowFragment.graphql";
import { useFragment, useMutation } from "react-relay";
import {
  Button,
  IconArrowInbox,
  IconLock,
  IconPageTextLine,
  Spinner,
  useToast,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { TrustCenterFileRowDownloadMutation } from "./__generated__/TrustCenterFileRowDownloadMutation.graphql";
import { useMutationWithToasts } from "/hooks/useMutationWithToast";
import { downloadFile, formatError } from "@probo/helpers";
import { use, useState } from "react";
import { Viewer } from "/providers/Viewer";
import { MagicLinkDialog } from "./MagicLinkDialog";
import type { TrustCenterFileRow_requestAccessMutation } from "./__generated__/TrustCenterFileRow_requestAccessMutation.graphql";

const requestAccessMutation = graphql`
  mutation TrustCenterFileRow_requestAccessMutation(
    $input: RequestTrustCenterFileAccessInput!
  ) {
    requestTrustCenterFileAccess(input: $input) {
      trustCenterAccess {
        id
      }
    }
  }
`;

const downloadMutation = graphql`
  mutation TrustCenterFileRowDownloadMutation(
    $input: ExportTrustCenterFileInput!
  ) {
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

export function TrustCenterFileRow(props: {
  file: TrustCenterFileRowFragment$key;
}) {
  const { __ } = useTranslate();
  const viewer = use(Viewer);
  const { toast } = useToast();

  const file = useFragment(trustCenterFileRowFragment, props.file);
  const [hasRequested, setHasRequested] = useState(file.hasUserRequestedAccess);

  const [requestAccess, isRequestingAccess] =
    useMutation<TrustCenterFileRow_requestAccessMutation>(
      requestAccessMutation,
    );
  const [commitDownload, downloading] =
    useMutationWithToasts<TrustCenterFileRowDownloadMutation>(downloadMutation);

  const handleRequestAccess = () => {
    requestAccess({
      variables: {
        input: {
          trustCenterFileId: file.id,
        },
      },
      onCompleted: (_, errors) => {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(__("Cannot request access"), errors),
            variant: "error",
          });
          return;
        }
        setHasRequested(true);
        toast({
          title: __("Success"),
          description: __("Access request submitted successfully."),
          variant: "success",
        });
      },
      onError: (error) => {
        toast({
          title: __("Error"),
          description: error.message ?? __("Cannot request access"),
          variant: "error",
        });
      },
    });
  };

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

  return (
    <div className="text-sm border border-border-solid -mt-px flex gap-3 flex-col md:flex-row md:justify-between px-6 py-3">
      <div className="flex items-center gap-2">
        <IconPageTextLine size={16} className=" flex-none text-txt-tertiary" />
        {file.name}
      </div>
      {!viewer && (
        <MagicLinkDialog>
          <Button
            className="w-full md:w-max"
            variant="secondary"
            icon={IconLock}
          >
            {hasRequested ? __("Access requested") : __("Request access")}
          </Button>
        </MagicLinkDialog>
      )}
      {viewer &&
        (file.isUserAuthorized ? (
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
          <Button
            disabled={hasRequested || isRequestingAccess}
            className="w-full md:w-max"
            variant="secondary"
            icon={IconLock}
            onClick={handleRequestAccess}
          >
            {hasRequested ? __("Access requested") : __("Request access")}
          </Button>
        ))}
    </div>
  );
}
