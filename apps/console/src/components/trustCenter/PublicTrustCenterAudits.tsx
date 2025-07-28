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
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { sprintf } from "@probo/helpers";
import { FrameworkLogo } from "/components/FrameworkLogo";

type Audit = {
  id: string;
  framework: {
    name: string;
  };
  validFrom: string;
  validUntil: string | null;
  state: string;
  createdAt: string;
  report: {
    id: string;
    filename: string;
    downloadUrl: string | null;
  } | null;
  reportUrl: string | null;
};

type Props = {
  audits: Audit[];
  organizationName: string;
};

export function PublicTrustCenterAudits({ audits, organizationName }: Props) {
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
            const hasReport = audit.report || audit.reportUrl;
            const downloadUrl = audit.report?.downloadUrl || audit.reportUrl;
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
                  {hasReport && downloadUrl ? (
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
