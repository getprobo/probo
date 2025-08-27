import {
  Card,
  Tr,
  Td,
  Table,
  Thead,
  Tbody,
  Th,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { faviconUrl, sprintf } from "@probo/helpers";
import type { TrustCenterVendor } from "../pages/PublicTrustCenterPage";

type Props = {
  vendors: TrustCenterVendor[];
  organizationName: string;
};

export function PublicTrustCenterVendors({ vendors, organizationName }: Props) {
  const { __ } = useTranslate();

  if (vendors.length === 0) {
    return (
      <Card padded>
        <div className="text-center py-8">
          <h2 className="text-xl font-semibold text-txt-primary mb-2">
            {__("Subcontractors")}
          </h2>
          <p className="text-txt-secondary">
            {__("No subcontractor information is currently available.")}
          </p>
        </div>
      </Card>
    );
  }

  return (
    <Card padded className="space-y-4">
      <div>
        <h2 className="text-xl font-semibold text-txt-primary">
          {__("Subcontractors")}
        </h2>
        <p className="text-sm text-txt-secondary mt-1">
          {sprintf(__("Third-party subcontractors %s work with"), organizationName)}
        </p>
      </div>

      <Table>
        <Thead>
          <Tr>
            <Th>{__("Company")}</Th>
            <Th>{__("Website")}</Th>
          </Tr>
        </Thead>
        <Tbody>
          {vendors.map((vendor) => {
            const url = vendor.privacyPolicyUrl || vendor.websiteUrl;
            const logo = faviconUrl(vendor.websiteUrl);

            const getCleanUrl = (url: string) => {
              try {
                const parsedUrl = new URL(url);
                return parsedUrl.hostname + parsedUrl.pathname + parsedUrl.search;
              } catch {
                return url.replace(/^https?:\/\//, '');
              }
            };

            return (
              <Tr key={vendor.id}>
                <Td>
                  <div className="flex items-center space-x-3">
                    {logo && (
                      <img
                        src={logo}
                        alt={`${vendor.name} logo`}
                        className="w-8 h-8 object-contain rounded-full"
                      />
                    )}
                    <div className="font-medium">
                      {vendor.name}
                    </div>
                  </div>
                </Td>
                <Td>
                  {url ? (
                    <a
                      href={url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-txt-info hover:opacity-80 underline transition-opacity"
                    >
                      {getCleanUrl(url)}
                    </a>
                  ) : (
                    <span className="text-txt-secondary text-sm">
                      {__("No website available")}
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
