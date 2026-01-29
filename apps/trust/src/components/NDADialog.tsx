import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, Card, Field, Logo, Spinner } from "@probo/ui";
import { clsx } from "clsx";
import { use, useEffect } from "react";
import { graphql } from "relay-runtime";
import { useWindowSize } from "usehooks-ts";
import { z } from "zod";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToast";
import { Viewer } from "#/providers/Viewer";

import { PDFPreview } from "./PDFPreview";

const signMutation = graphql`
  mutation NDADialogSignMutation($input: AcceptNonDisclosureAgreementInput!) {
    acceptNonDisclosureAgreement(input: $input) {
      success
    }
  }
`;

const schema = z.object({
  fullName: z.string(),
});

export function NDADialog({
  organizationName,
  url,
  fileName,
}: {
  organizationName: string;
  url?: string | null;
  fileName?: string | null;
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
  const viewer = use(Viewer);

  const {
    handleSubmit: handleSubmitWrapper,
    register,
    formState,
  } = useFormWithSchema(schema, {
    defaultValues: {
      fullName: viewer?.fullName,
    },
  });

  const [commitSigning, isSigning] = useMutationWithToasts(signMutation, {
    onSuccess: () => {
      window.location.reload();
    },
  });

  const handleSubmit = handleSubmitWrapper(async ({ fullName }) => {
    await commitSigning({
      variables: {
        input: {
          fullName,
        },
      },
    });
  });

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
                "Access to %s compliance page documents requires signing a Non-Disclosure Agreement (NDA). Please review the agreement below. Once signed, you’ll receive immediate access to the requested documents.",
              ),
              organizationName,
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
          <form onSubmit={e => void handleSubmit(e)}>
            <div className="mt-4">
              <Field
                required
                label={__("Full name")}
                placeholder="John Doe"
                {...register("fullName")}
                type="text"
              />
              <Button
                type="submit"
                className="h-10 w-full my-8"
                disabled={formState.isSubmitting || !formState.isValid}
                icon={isSigning ? Spinner : undefined}
              >
                {__("Review & Sign")}
              </Button>
            </div>
          </form>
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
            Powered by
            {" "}
            <Logo withPicto className="h-6" />
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
