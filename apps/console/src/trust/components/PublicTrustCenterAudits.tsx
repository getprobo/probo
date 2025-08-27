import {
  Card,
  Tr,
  Td,
  Table,
  Thead,
  Tbody,
  Th,
  Button,
  IconArrowDown,
  IconLock,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { sprintf } from "@probo/helpers";
import { FrameworkLogo } from "/components/FrameworkLogo";
import { PublicTrustCenterAccessRequestDialog } from "./PublicTrustCenterAccessRequestDialog";
import type { TrustCenterAudit } from "../pages/PublicTrustCenterPage";

type Props = {
  audits: TrustCenterAudit[];
  organizationName: string;
  isAuthenticated: boolean;
  trustCenterId: string;
};

export function PublicTrustCenterAudits({
  audits,
  organizationName,
  isAuthenticated,
  trustCenterId
}: Props) {
  const { __ } = useTranslate();

  if (audits.length === 0) {
    return (
      <Card padded>
        <div className="text-center py-8">
          <h2 className="text-xl font-semibold text-txt-primary mb-2">
            {__("Compliance")}
          </h2>
          <p className="text-txt-secondary">
            {__("No compliance reports are currently available.")}
          </p>
        </div>
      </Card>
    );
  }

  return (
    <Card padded className="space-y-4">
      <div>
        <h2 className="text-xl font-semibold text-txt-primary">
          {__("Compliance")}
        </h2>
        <p className="text-sm text-txt-secondary mt-1">
          {sprintf(__("%s is compliant with the following frameworks"), organizationName)}
        </p>
      </div>

      <Table>
        <Thead>
          <Tr>
            <Th>{__("Framework")}</Th>
            <Th>{__("Report")}</Th>
          </Tr>
        </Thead>
        <Tbody>
          {audits.map((audit) => {
            const hasReport = audit.report !== null;
            const downloadUrl = audit.report?.downloadUrl;
            const reportName = audit.report?.filename || __("Compliance Report");

            return (
              <Tr key={audit.id}>
                <Td>
                  <div className="flex items-center gap-3">
                    <div className="flex-shrink-0">
                      <div className="w-8 h-8 [&>img]:w-8 [&>img]:h-8 [&>div]:w-8 [&>div]:h-8">
                        <FrameworkLogo name={audit.framework.name} />
                      </div>
                    </div>
                    <div className="font-medium">
                      {audit.framework.name}
                    </div>
                  </div>
                </Td>
                <Td>
                  {!hasReport ? (
                    <span className="text-txt-tertiary text-sm">
                      {__("No report")}
                    </span>
                  ) : !isAuthenticated ? (
                    <PublicTrustCenterAccessRequestDialog
                      trigger={
                        <Button
                          variant="secondary"
                          icon={IconLock}
                        >
                          {__("Request Access")}
                        </Button>
                      }
                      trustCenterId={trustCenterId}
                      organizationName={organizationName}
                    />
                  ) : downloadUrl ? (
                    <Button
                      variant="secondary"
                      icon={IconArrowDown}
                      onClick={() => {
                        const link = document.createElement('a');
                        link.href = downloadUrl;
                        link.download = reportName;
                        document.body.appendChild(link);
                        link.click();
                        document.body.removeChild(link);
                      }}
                    >
                      {__("Download")}
                    </Button>
                  ) : (
                    <span className="text-txt-tertiary text-sm">
                      {__("Not available")}
                    </span>
                  )}
                </Td>
              </Tr>
            );
          })}
        </Tbody>
      </Table>
    </Card>
  );
}
