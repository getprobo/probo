import { downloadFile, formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  FrameworkLogo,
  IconArrowInbox,
  IconLock,
  IconMedal,
  Spinner,
  Table,
  useToast,
} from "@probo/ui";
import { type PropsWithChildren, use, useState } from "react";
import { useFragment, useMutation } from "react-relay";
import { useLocation } from "react-router";
import { graphql } from "relay-runtime";

import { useMutationWithToasts } from "#/hooks/useMutationWithToast";
import { Viewer } from "#/providers/Viewer";

import type { AuditRow_requestAccessMutation } from "./__generated__/AuditRow_requestAccessMutation.graphql";
import type { AuditRowDownloadMutation } from "./__generated__/AuditRowDownloadMutation.graphql";
import type { AuditRowFragment$key } from "./__generated__/AuditRowFragment.graphql";

const requestAccessMutation = graphql`
  mutation AuditRow_requestAccessMutation($input: RequestReportAccessInput!) {
    requestReportAccess(input: $input) {
      trustCenterAccess {
        id
      }
    }
  }
`;

const downloadMutation = graphql`
  mutation AuditRowDownloadMutation($input: ExportReportPDFInput!) {
    exportReportPDF(input: $input) {
      data
    }
  }
`;

const auditRowFragment = graphql`
  fragment AuditRowFragment on Audit {
    report {
      id
      filename
      isUserAuthorized
      hasUserRequestedAccess
    }
    framework {
      id
      name
      lightLogoURL
      darkLogoURL
    }
  }
`;

export function AuditRow(props: { audit: AuditRowFragment$key }) {
  const { __ } = useTranslate();
  const viewer = use(Viewer);
  const { toast } = useToast();

  const audit = useFragment(auditRowFragment, props.audit);
  const [hasRequested, setHasRequested] = useState(
    audit.report?.hasUserRequestedAccess,
  );

  const [requestAccess, isRequestingAccess]
    = useMutation<AuditRow_requestAccessMutation>(requestAccessMutation);
  const [commitDownload, downloading]
    = useMutationWithToasts<AuditRowDownloadMutation>(downloadMutation);

  const handleRequestAccess = () => {
    requestAccess({
      variables: {
        input: {
          reportId: audit.report?.id ?? "",
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
    if (!audit.report?.id) {
      return;
    }
    await commitDownload({
      variables: {
        input: {
          reportId: audit.report.id,
        },
      },
      onSuccess(response) {
        downloadFile(response.exportReportPDF.data, audit.report!.filename);
      },
    });
  };

  return (
    <div className="text-sm border border-border-solid -mt-px flex gap-3 flex-col md:flex-row md:justify-between px-6 py-3">
      <div className="flex items-center gap-2">
        <IconMedal size={16} className="flex-none text-txt-tertiary" />
        {audit.framework.name}
      </div>
      {audit.report && audit.report.isUserAuthorized
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

export function AuditRowAvatar(props: { audit: AuditRowFragment$key }) {
  const audit = useFragment(auditRowFragment, props.audit);

  return (
    <>
      <AuditDialog audit={props.audit}>
        <button
          className="block cursor-pointer aspect-square"
          title={`Logo ${audit.framework.name}`}
        >
          <div className="flex flex-col gap-2 items-center w-19">
            <FrameworkLogo
              className="size-19"
              lightLogoURL={audit.framework.lightLogoURL}
              darkLogoURL={audit.framework.darkLogoURL}
              name={audit.framework.name}
            />
            <div className="txt-primary text-sm max-w-19 overflow-hidden min-w-0 whitespace-nowrap text-ellipsis">
              {audit.framework.name}
            </div>
          </div>
        </button>
      </AuditDialog>
    </>
  );
}

function AuditDialog(
  props: PropsWithChildren<{ audit: AuditRowFragment$key; logo?: string }>,
) {
  const audit = useFragment(auditRowFragment, props.audit);
  const location = useLocation();
  const { __ } = useTranslate();
  const items = [
    {
      label: __("Certifications"),
      to: location.pathname,
    },
    {
      label: audit.framework.name,
      to: location.pathname,
    },
  ];
  return (
    <Dialog
      trigger={props.children}
      className="max-w-[500px]"
      title={<Breadcrumb items={items} />}
    >
      <DialogContent className="p-4 lg:p-8 space-y-6">
        <FrameworkLogo
          className="size-24 mx-auto"
          lightLogoURL={audit.framework.lightLogoURL}
          darkLogoURL={audit.framework.darkLogoURL}
          name={audit.framework.name}
        />
        <h2 className="text-xl font-semibold mb-1">{audit.framework.name}</h2>
        <Table>
          <AuditRow audit={props.audit} />
        </Table>
      </DialogContent>
    </Dialog>
  );
}
