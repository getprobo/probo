import { useCopy } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, DialogContent } from "@probo/ui";

export function SAMLDomainVerifyDialog(props: {
  domainVerificationToken: string;
}) {
  const { domainVerificationToken } = props;

  const { __ } = useTranslate();

  const dnsRecord = `probo-verification=${domainVerificationToken}`;
  const [isCopied, copy] = useCopy();

  return (
    <>
      <DialogContent padded className="space-y-6">
        <div>
          <h3 className="text-base font-medium mb-4">
            {__("Verify Domain Ownership")}
          </h3>
          <p className="text-sm text-gray-600 mb-4">
            {__(
              "Add the following TXT record to your domain's DNS configuration to verify ownership:",
            )}
          </p>
          <div className="bg-gray-50 rounded-lg p-4 mb-4">
            <div className="space-y-2">
              <div>
                <span className="font-semibold text-sm">
                  {__("Host/Name:")}
                </span>
                <code className="ml-2 bg-white px-2 py-1 rounded text-sm">
                  @
                </code>
                <span className="ml-2 text-xs text-gray-600">
                  {__("or use your domain name")}
                </span>
              </div>
              <div>
                <span className="font-semibold text-sm">{__("Type:")}</span>
                <code className="ml-2 bg-white px-2 py-1 rounded text-sm">
                  TXT
                </code>
              </div>
              <div>
                <span className="font-semibold text-sm">{__("Value:")}</span>
                <div className="mt-1 flex items-center gap-2">
                  <code className="flex-1 bg-white px-2 py-1 rounded text-sm break-all font-mono">
                    {dnsRecord}
                  </code>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => copy(dnsRecord)}
                  >
                    {isCopied ? __("Copied!") : __("Copy")}
                  </Button>
                </div>
              </div>
            </div>
          </div>
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <p className="text-sm text-blue-800">
              <strong>{__("Note:")}</strong>
              {" "}
              {__(
                "DNS changes may take up to 48 hours to propagate, but typically complete within a few minutes.",
              )}
            </p>
          </div>
        </div>
      </DialogContent>
    </>
  );
}
