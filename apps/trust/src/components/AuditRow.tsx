import { graphql } from "relay-runtime";
import type { AuditRowFragment$key } from "./__generated__/AuditRowFragment.graphql";
import { useFragment, useMutation } from "react-relay";
import {
  Button,
  FrameworkLogo,
  IconArrowInbox,
  IconInboxEmpty,
  IconLock,
  IconMedal,
  Spinner,
  Td,
  Tr,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useIsAuthenticated } from "/hooks/useIsAuthenticated";
import type { AuditRowDownloadMutation } from "./__generated__/AuditRowDownloadMutation.graphql";
import { useMutationWithToasts } from "/hooks/useMutationWithToast";
import { downloadFile } from "@probo/helpers";

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
    <div>
      <FrameworkLogo
        alt={audit.framework.name}
        name={audit.framework.name}
        className="size-18"
      />
    </div>
  );
}
