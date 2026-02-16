import { downloadFile, formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  IconArrowInbox,
  IconLock,
  IconPageTextLine,
  Spinner,
  useToast,
} from "@probo/ui";
import { use, useState } from "react";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import { useMutationWithToasts } from "#/hooks/useMutationWithToast";
import { Viewer } from "#/providers/Viewer";

import type { DocumentRow_requestAccessMutation } from "./__generated__/DocumentRow_requestAccessMutation.graphql";
import type { DocumentRowDownloadMutation } from "./__generated__/DocumentRowDownloadMutation.graphql";
import type { DocumentRowFragment$key } from "./__generated__/DocumentRowFragment.graphql";

const requestAccessMutation = graphql`
  mutation DocumentRow_requestAccessMutation(
    $input: RequestDocumentAccessInput!
  ) {
    requestDocumentAccess(input: $input) {
      trustCenterAccess {
        id
      }
    }
  }
`;

const downloadMutation = graphql`
  mutation DocumentRowDownloadMutation($input: ExportDocumentPDFInput!) {
    exportDocumentPDF(input: $input) {
      data
    }
  }
`;

const documentRowFragment = graphql`
  fragment DocumentRowFragment on Document {
    id
    title
    isUserAuthorized
    hasUserRequestedAccess
  }
`;

export function DocumentRow(props: { document: DocumentRowFragment$key }) {
  const { __ } = useTranslate();
  const viewer = use(Viewer);
  const { toast } = useToast();

  const document = useFragment(documentRowFragment, props.document);
  const [hasRequested, setHasRequested] = useState(
    document.hasUserRequestedAccess,
  );

  const [requestAccess, isRequestingAccess]
    = useMutation<DocumentRow_requestAccessMutation>(requestAccessMutation);
  const [commitDownload, downloading]
    = useMutationWithToasts<DocumentRowDownloadMutation>(downloadMutation);

  const handleRequestAccess = () => {
    requestAccess({
      variables: {
        input: {
          documentId: document.id,
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

  const handleDownload = async () => {
    await commitDownload({
      variables: {
        input: {
          documentId: document.id,
        },
      },
      onSuccess(response) {
        downloadFile(response.exportDocumentPDF.data, document.title);
      },
    });
  };

  return (
    <div className="text-sm border border-border-solid -mt-px flex gap-3 flex-col md:flex-row md:justify-between px-6 py-3">
      <div className="flex items-center gap-2">
        <IconPageTextLine size={16} className=" flex-none text-txt-tertiary" />
        {document.title}
      </div>
      {document.isUserAuthorized
        ? (
            <Button
              className="w-full md:w-max"
              variant="secondary"
              disabled={downloading}
              icon={downloading ? Spinner : IconArrowInbox}
              onClick={() => void handleDownload()}
            >
              {downloading ? __("Downloading") : __("Download")}
            </Button>
          )
        : viewer
          ? (
              <Button
                disabled={hasRequested || isRequestingAccess}
                className="w-full md:w-max"
                variant="secondary"
                icon={IconLock}
                onClick={handleRequestAccess}
              >
                {hasRequested ? __("Access requested") : __("Request access")}
              </Button>
            )
          : (
              <Button
                className="w-full md:w-max"
                variant="secondary"
                icon={IconLock}
                to="/connect"
              >
                {hasRequested ? __("Access requested") : __("Request access")}
              </Button>
            )}
    </div>
  );
}
