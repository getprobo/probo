import { Button, Card, Logo, Spinner } from "@probo/ui";
import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useEffect } from "react";
import { PDFPreview } from "./PDFPreview";
import { useWindowSize } from "usehooks-ts";
import clsx from "clsx";
import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "/hooks/useMutationWithToast";

const signMutation = graphql`
  mutation NDADialogSignMutation($input: AcceptNonDisclosureAgreementInput!) {
    acceptNonDisclosureAgreement(input: $input) {
      success
    }
  }
`;

export function NDADialog({
  name,
  url,
  fileName,
  trustCenterId,
}: {
  name: string;
  url?: string | null;
  fileName?: string | null;
  trustCenterId: string;
}) {
  const { __ } = useTranslate();
  useEffect(() => {
    document.body.style.setProperty("overflow", "hidden");
    return () => {
      document.body.style.removeProperty("overflow");
    };
  }, []);
  const { width } = useWindowSize();
  const isMobile = width < 1100;
  const isDesktop = !isMobile;
  const [commitSigning, isSigning] = useMutationWithToasts(signMutation, {
    onSuccess: () => {
      window.location.reload();
    },
  });

  const handleSign = () => {
    commitSigning({
      variables: {
        input: {
          trustCenterId,
        },
      },
    });
  };

  return (
    <div className="fixed inset-0 bg-level-2 z-100 flex flex-col lg:h-screen">
      <header className="flex items-center h-12 justify-between border-b border-border-solid px-4 flex-none">
        <Logo />
      </header>
      <div className="grid lg:grid-cols-2 min-h-0 h-full">
        <div className="max-w-[440px] mx-auto py-20">
          <h1 className="text-2xl font-semibold mb-4">
            {__("Review & Sign NDA")}
          </h1>
          <p className="text-txt-secondary">
            {sprintf(
              __(
                "Access to %s Trust Center documents requires signing a Non-Disclosure Agreement (NDA). Please review the agreement below. Once signed, you’ll receive immediate access to the requested documents.",
              ),
              name,
            )}
          </p>
          {isMobile && url && (
            <Card className="flex justify-between py-3 px-4 text-sm items-center my-6">
              {fileName}
              <Button variant="secondary" asChild>
                <a target="_blank" rel="noopener noreferrer" href={url}>
                  {__("View document")}
                </a>
              </Button>
            </Card>
          )}
          <Button
            onClick={handleSign}
            className="h-10 w-full my-8"
            disabled={isSigning}
            icon={isSigning ? Spinner : undefined}
          >
            {__("Review & Sign")}
          </Button>
          <p className="text-xs text-txt-secondary">
            {__(
              "By clicking Review & Sign, you agree to the terms of this NDA. If you have questions about the NDA, please contact security@probo.com.",
            )}
          </p>
          <a
            href="https://www.getprobo.com/"
            className={clsx(
              "flex gap-1 text-sm font-medium text-txt-tertiary items-center w-max mx-auto",
              isMobile ? "mt-15" : "mt-30",
            )}
          >
            Powered by <Logo withPicto className="h-6" />
          </a>
        </div>
        {isDesktop && (
          <div className="bg-subtle h-full border-l border-border-solid min-h-0">
            {url && <PDFPreview src={url} name={fileName ?? ""} />}
          </div>
        )}
      </div>
    </div>
  );
}
