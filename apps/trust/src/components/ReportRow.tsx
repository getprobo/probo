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

import type { ReportRow_requestAccessMutation } from "./__generated__/ReportRow_requestAccessMutation.graphql";
import type { ReportRowDownloadMutation } from "./__generated__/ReportRowDownloadMutation.graphql";
import type { ReportRowFragment$key } from "./__generated__/ReportRowFragment.graphql";

const requestAccessMutation = graphql`
  mutation ReportRow_requestAccessMutation($input: RequestReportAccessInput!) {
    requestReportAccess(input: $input) {
      trustCenterAccess {
        id
      }
    }
  }
`;

const downloadMutation = graphql`
  mutation ReportRowDownloadMutation($input: ExportReportPDFInput!) {
    exportReportPDF(input: $input) {
      data
    }
  }
`;

const reportRowFragment = graphql`
  fragment ReportRowFragment on Report {
    id
    frameworkType
    file {
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

export function ReportRow(props: { report: ReportRowFragment$key }) {
  const { __ } = useTranslate();
  const viewer = use(Viewer);
  const { toast } = useToast();

  const report = useFragment(reportRowFragment, props.report);
  const [hasRequested, setHasRequested] = useState(
    report.file?.hasUserRequestedAccess,
  );

  const [requestAccess, isRequestingAccess]
    = useMutation<ReportRow_requestAccessMutation>(requestAccessMutation);
  const [commitDownload, downloading]
    = useMutationWithToasts<ReportRowDownloadMutation>(downloadMutation);

  const handleRequestAccess = () => {
    requestAccess({
      variables: {
        input: {
          reportId: report.id,
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
    if (!report.file) {
      return;
    }
    await commitDownload({
      variables: {
        input: {
          reportId: report.id,
        },
      },
      onSuccess(response) {
        downloadFile(response.exportReportPDF.data, report.file!.filename);
      },
    });
  };

  return (
    <div className="text-sm border border-border-solid -mt-px flex gap-3 flex-col md:flex-row md:justify-between px-6 py-3">
      <div className="flex items-center gap-2">
        <IconMedal size={16} className="flex-none text-txt-tertiary" />
        <span>{report.framework.name}</span>
        {report.frameworkType && (
          <span className="text-sm italic text-txt-tertiary">{report.frameworkType}</span>
        )}
      </div>
      {report.file && report.file.isUserAuthorized
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

export function ReportRowAvatar(props: { report: ReportRowFragment$key }) {
  const report = useFragment(reportRowFragment, props.report);

  return (
    <>
      <ReportDialog report={props.report}>
        <button
          className="block cursor-pointer aspect-square"
          title={`Logo ${report.framework.name}`}
        >
          <div className="flex flex-col gap-2 items-center w-19">
            <FrameworkLogo
              className="size-19"
              lightLogoURL={report.framework.lightLogoURL}
              darkLogoURL={report.framework.darkLogoURL}
              name={report.framework.name}
            />
            <div className="txt-primary text-sm max-w-19 overflow-hidden min-w-0 whitespace-nowrap text-ellipsis">
              {report.framework.name}
            </div>
          </div>
        </button>
      </ReportDialog>
    </>
  );
}

function ReportDialog(
  props: PropsWithChildren<{ report: ReportRowFragment$key; logo?: string }>,
) {
  const report = useFragment(reportRowFragment, props.report);
  const location = useLocation();
  const { __ } = useTranslate();
  const items = [
    {
      label: __("Certifications"),
      to: location.pathname,
    },
    {
      label: report.framework.name,
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
          lightLogoURL={report.framework.lightLogoURL}
          darkLogoURL={report.framework.darkLogoURL}
          name={report.framework.name}
        />
        <h2 className="text-xl font-semibold mb-1">{report.framework.name}</h2>
        <Table>
          <ReportRow report={props.report} />
        </Table>
      </DialogContent>
    </Dialog>
  );
}
