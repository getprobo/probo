import { graphql } from "relay-runtime";
import type { AuditRowFragment$key } from "./__generated__/AuditRowFragment.graphql";
import { useFragment } from "react-relay";
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
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { AuditRowDownloadMutation } from "./__generated__/AuditRowDownloadMutation.graphql";
import { useMutationWithToasts } from "/hooks/useMutationWithToast";
import { downloadFile } from "@probo/helpers";
import { type PropsWithChildren, useState } from "react";
import { useLocation } from "react-router";
import { RequestAccessDialog } from "/components/RequestAccessDialog.tsx";

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
  const audit = useFragment(auditRowFragment, props.audit);
  const { __ } = useTranslate();
  const [commitDownload, downloading] =
    useMutationWithToasts<AuditRowDownloadMutation>(downloadMutation);
  const handleDownload = () => {
    if (!audit.report?.id) {
      return;
    }
    commitDownload({
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

  const [hasRequested, setHasRequested] = useState(
    audit.report?.hasUserRequestedAccess
  );
  return (
    <div className="text-sm border border-border-solid -mt-px flex gap-3 flex-col md:flex-row md:justify-between px-6 py-3">
      <div className="flex items-center gap-2">
        <IconMedal size={16} className="flex-none text-txt-tertiary" />
        {audit.framework.name}
      </div>
      {audit.report && audit.report.isUserAuthorized && (
        <Button
          className="w-full md:w-max"
          variant="secondary"
          disabled={downloading}
          icon={downloading ? Spinner : IconArrowInbox}
          onClick={handleDownload}
        >
          {__("Download")}
        </Button>
      )}
      {audit.report && !audit.report.isUserAuthorized && (
        <RequestAccessDialog
          reportId={audit.report.id}
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

export function AuditRowAvatar(props: { audit: AuditRowFragment$key }) {
  const audit = useFragment(auditRowFragment, props.audit);

  return (
    <>
      <AuditDialog audit={props.audit}>
        <button
          className="block cursor-pointer aspect-square"
          title={`Logo ${audit.framework.name}`}
        >
          <FrameworkLogo
            className="size-19"
            lightLogoURL={audit.framework.lightLogoURL}
            darkLogoURL={audit.framework.darkLogoURL}
            name={audit.framework.name}
          />
        </button>
      </AuditDialog>
    </>
  );
}

function AuditDialog(
  props: PropsWithChildren<{ audit: AuditRowFragment$key; logo?: string }>
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
