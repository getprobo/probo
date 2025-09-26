import { graphql } from "relay-runtime";
import type {
  AuditRowFragment$data,
  AuditRowFragment$key,
} from "./__generated__/AuditRowFragment.graphql";
import { useFragment, useMutation } from "react-relay";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  FrameworkLogo,
  IconArrowInbox,
  IconInboxEmpty,
  IconLock,
  IconMedal,
  Spinner,
  Table,
  Td,
  Tr,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useIsAuthenticated } from "/hooks/useIsAuthenticated";
import type { AuditRowDownloadMutation } from "./__generated__/AuditRowDownloadMutation.graphql";
import { useMutationWithToasts } from "/hooks/useMutationWithToast";
import { downloadFile } from "@probo/helpers";
import type { PropsWithChildren } from "react";
import { useLocation, useNavigation } from "react-router";

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
  const isAuthenticated = useIsAuthenticated();
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
  return (
    <Tr className="text-sm *:border-border-solid *:border-b-1">
      <Td>
        <div className="flex items-center gap-2">
          <IconMedal size={16} />
          {audit.framework.name}
        </div>
      </Td>
      <Td className="text-end">
        {audit.report ? (
          isAuthenticated ? (
            <Button
              className="ml-auto"
              variant="secondary"
              disabled={downloading}
              icon={downloading ? Spinner : IconArrowInbox}
              onClick={handleDownload}
            >
              {__("Download")}
            </Button>
          ) : (
            <Button className="ml-auto" variant="secondary" icon={IconLock}>
              {__("Request access")}
            </Button>
          )
        ) : (
          <span className=" text-txt-secondary">{__("No report")}</span>
        )}
      </Td>
    </Tr>
  );
}

export function AuditRowAvatar(props: { audit: AuditRowFragment$key }) {
  const audit = useFragment(auditRowFragment, props.audit);
  return (
    <AuditDialog audit={props.audit}>
      <button className="block cursor-pointer">
        <FrameworkLogo
          alt={audit.framework.name}
          name={audit.framework.name}
          className="size-18 block"
        />
      </button>
    </AuditDialog>
  );
}

function AuditDialog(
  props: PropsWithChildren<{ audit: AuditRowFragment$key }>
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
          alt={audit.framework.name}
          name={audit.framework.name}
          className="size-24 block mx-auto"
        />
        <h2 className="text-xl font-semibold mb-1">{audit.framework.name}</h2>
        <p className="text-txt-secondary">Framework description</p>
        <Table>
          <AuditRow audit={props.audit} />
        </Table>
      </DialogContent>
    </Dialog>
  );
}
