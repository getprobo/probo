import { graphql } from "relay-runtime";
import type { AuditRowFragment$key } from "./__generated__/AuditRowFragment.graphql";
import { useFragment } from "react-relay";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  IconArrowInbox,
  IconLock,
  IconMedal,
  Spinner,
  Table,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { AuditRowDownloadMutation } from "./__generated__/AuditRowDownloadMutation.graphql";
import { useMutationWithToasts } from "/hooks/useMutationWithToast";
import { downloadFile, getLogoUrl } from "@probo/helpers";
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
    audit.report?.hasUserRequestedAccess,
  );
  return (
    <div className="text-sm border-1 border-border-solid -mt-[1px] flex gap-3 flex-col md:flex-row md:justify-between px-6 py-3">
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

  const logos = {
    "ISO 27001 (2022)": getLogoUrl("iso27001.svg"),
    "SOC 2": getLogoUrl("soc2.svg"),
    HIPAA: getLogoUrl("hipaa.svg"),
    GDPR: getLogoUrl("gdpr.svg"),
  };

  return (
    <>
      <AuditDialog
        audit={props.audit}
        logo={logos[audit.framework.name as keyof typeof logos]}
      >
        <button
          className="block cursor-pointer aspect-square"
          title={`Logo ${audit.framework.name}`}
        >
          {audit.framework.name in logos ? (
            <img
              src={logos[audit.framework.name as keyof typeof logos]}
              alt={`${audit.framework.name} Logo`}
            />
          ) : (
            <div
              className="bg-[#F0F7E2] aspect-square w-full rounded-full text-xs text-[#000] font-bold flex items-center justify-center pb-6 px-2"
              style={{ background: `url(${getLogoUrl("blank.svg")}) no-repeat` }}
            >
              <span className="line-clamp-2 overflow-hidden">
                {" "}
                {audit.framework.name}
              </span>
            </div>
          )}
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
        {props.logo && (
          <img
            alt={audit.framework.name}
            src={props.logo}
            className="size-24 block mx-auto"
          />
        )}
        <h2 className="text-xl font-semibold mb-1">{audit.framework.name}</h2>
        <Table>
          <AuditRow audit={props.audit} />
        </Table>
      </DialogContent>
    </Dialog>
  );
}
