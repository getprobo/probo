import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, Card, Field, IconCircleX, Logo, Spinner } from "@probo/ui";
import { startTransition, useEffect, useRef } from "react";
import {
  type PreloadedQuery,
  useMutation,
  usePreloadedQuery,
  useRefetchableFragment,
} from "react-relay";
import { Navigate, useNavigate } from "react-router";
import { graphql } from "relay-runtime";
import { useWindowSize } from "usehooks-ts";
import { z } from "zod";

import { PDFPreview } from "#/components/PDFPreview";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

import type { NDAPageAcceptElectronicSignatureMutation } from "./__generated__/NDAPageAcceptElectronicSignatureMutation.graphql";
import type { NDAPageFragment$key } from "./__generated__/NDAPageFragment.graphql";
import type { NDAPageQuery as NDAPageQueryType } from "./__generated__/NDAPageQuery.graphql";
import type { NDAPageRecordSigningEventMutation } from "./__generated__/NDAPageRecordSigningEventMutation.graphql";
import type { NDAPageRefetchQuery } from "./__generated__/NDAPageRefetchQuery.graphql";

export const ndaPageQuery = graphql`
  query NDAPageQuery {
    viewer {
      fullName
    }
    currentTrustCenter @required(action: THROW) {
      organization {
        name
      }
      nonDisclosureAgreement {
        fileName
        fileUrl
        viewerSignature {
          status
        }
      }
      ...NDAPageFragment
    }
  }
`;

const ndaPageFragment = graphql`
  fragment NDAPageFragment on TrustCenter
  @refetchable(queryName: "NDAPageRefetchQuery") {
    nonDisclosureAgreement @required(action: THROW) {
      viewerSignature {
        id
        status
        consentText
        lastError
      }
    }
  }
`;

const acceptElectronicSignatureMutation = graphql`
  mutation NDAPageAcceptElectronicSignatureMutation(
    $input: AcceptElectronicSignatureInput!
  ) {
    acceptElectronicSignature(input: $input) {
      signature {
        id
        status
      }
    }
  }
`;

const recordSigningEventMutation = graphql`
  mutation NDAPageRecordSigningEventMutation(
    $input: RecordSigningEventInput!
  ) {
    recordSigningEvent(input: $input) {
      success
    }
  }
`;

const schema = z.object({
  fullName: z.string().min(1),
});

export function NDAPage(props: {
  queryRef: PreloadedQuery<NDAPageQueryType>;
}) {
  const { __ } = useTranslate();
  const navigate = useNavigate();
  const documentViewedRef = useRef(false);

  const queryData = usePreloadedQuery(ndaPageQuery, props.queryRef);
  const trustCenter = queryData.currentTrustCenter;
  const viewer = queryData.viewer;

  const [data, refetch] = useRefetchableFragment<NDAPageRefetchQuery, NDAPageFragment$key>(
    ndaPageFragment,
    trustCenter,
  );
  const ndaSignature = data.nonDisclosureAgreement.viewerSignature;

  const { width } = useWindowSize();
  const isMobile = width < 1100;
  const isDesktop = !isMobile;

  const {
    handleSubmit: handleSubmitWrapper,
    register,
    formState,
  } = useFormWithSchema(schema, {
    defaultValues: {
      fullName: viewer?.fullName,
    },
  });

  const [acceptSignature, isAccepting] = useMutation<NDAPageAcceptElectronicSignatureMutation>(
    acceptElectronicSignatureMutation,
  );

  const [recordSigningEvent] = useMutation<NDAPageRecordSigningEventMutation>(
    recordSigningEventMutation,
  );

  const isProcessing
    = ndaSignature?.status === "ACCEPTED"
      || ndaSignature?.status === "PROCESSING";

  const isFailed = ndaSignature?.status === "FAILED";
  const isCompleted = ndaSignature?.status === "COMPLETED";

  useEffect(() => {
    if (isCompleted) {
      void navigate("/overview", { replace: true });
    }
  }, [isCompleted, navigate]);

  useEffect(() => {
    if (!isProcessing) return;

    const poll = () => startTransition(() => {
      refetch({}, { fetchPolicy: "network-only" });
    });
    const interval = setInterval(poll, 1500);

    return () => clearInterval(interval);
  }, [isProcessing, refetch]);

  useEffect(() => {
    if (
      ndaSignature
      && ndaSignature.status === "PENDING"
      && !documentViewedRef.current
    ) {
      documentViewedRef.current = true;
      recordSigningEvent({
        variables: {
          input: {
            signatureId: ndaSignature.id,
            eventType: "DOCUMENT_VIEWED",
          },
        },
      });
    }
  }, [ndaSignature, recordSigningEvent]);

  const handleSubmit = handleSubmitWrapper(({ fullName }) => {
    if (!ndaSignature) return;

    if (ndaSignature.status === "PENDING") {
      recordSigningEvent({
        variables: {
          input: {
            signatureId: ndaSignature.id,
            eventType: "FULL_NAME_TYPED",
          },
        },
      });
    }

    recordSigningEvent({
      variables: {
        input: {
          signatureId: ndaSignature.id,
          eventType: "CONSENT_GIVEN",
        },
      },
      onCompleted: () => {
        acceptSignature({
          variables: {
            input: {
              signatureId: ndaSignature.id,
              fullName,
            },
          },
        });
      },
    });
  });

  const nda = trustCenter.nonDisclosureAgreement;
  if (!viewer) {
    return <Navigate to="/connect" replace />;
  }

  if (!nda || !nda.viewerSignature || nda.viewerSignature.status === "COMPLETED") {
    return <Navigate to="/overview" replace />;
  }

  const consentText = ndaSignature?.consentText
    ? ndaSignature.consentText
    : __(
        "By clicking Review & Sign, you agree to the terms of this NDA. If you have questions about the NDA, please contact security@probo.com.",
      );

  return (
    <div className="bg-level-2 flex flex-col min-h-screen lg:h-screen">
      <header className="flex items-center h-12 justify-between border-b border-border-solid px-4 flex-none">
        <Logo />
      </header>
      <div className="grid lg:grid-cols-2 min-h-0 flex-1">
        <div className="flex flex-col items-center overflow-y-auto">
          <div className="max-w-[440px] w-full mx-auto px-4 py-12 lg:py-20 flex-1">
            <h1 className="text-2xl font-semibold">
              {__("Non-Disclosure Agreement")}
            </h1>
            <p className="text-txt-secondary mt-2">
              {sprintf(
                __(
                  "%s requires you to sign an NDA before accessing compliance documents.",
                ),
                trustCenter.organization.name,
              )}
            </p>
            {isMobile && nda?.fileUrl && (
              <Card className="flex justify-between py-3 px-4 text-sm items-center mt-6">
                {nda?.fileName}
                <Button variant="secondary" asChild>
                  <a target="_blank" rel="noopener noreferrer" href={nda.fileUrl}>
                    {__("View document")}
                  </a>
                </Button>
              </Card>
            )}
            <form
              onSubmit={e => void handleSubmit(e)}
              className="mt-8"
            >
              <Field
                required
                label={__("Full name")}
                placeholder="John Doe"
                {...register("fullName")}
                type="text"
                disabled={isProcessing}
              />
              <p className="text-xs text-txt-tertiary mt-6">
                {consentText}
              </p>
              {isFailed && (
                <div className="flex items-start gap-2 mt-4 rounded-md bg-red-50 border border-red-200 p-3 text-sm text-red-800">
                  <IconCircleX size={16} className="mt-0.5 shrink-0" />
                  <div>
                    <p className="font-medium">
                      {__("Signature processing failed")}
                    </p>
                    <p className="mt-0.5 text-red-600">
                      {ndaSignature?.lastError
                        ?? __("We encountered an issue processing your signature. Please try again.")}
                    </p>
                  </div>
                </div>
              )}
              {isProcessing
                ? (
                    <Button
                      type="button"
                      className="h-10 w-full mt-4"
                      disabled
                      icon={Spinner}
                    >
                      {__("Sealing your signature...")}
                    </Button>
                  )
                : (
                    <Button
                      type="submit"
                      className="h-10 w-full mt-4"
                      disabled={formState.isSubmitting || !formState.isValid}
                      icon={isAccepting ? Spinner : undefined}
                    >
                      {isFailed
                        ? __("Try again")
                        : __("Accept")}
                    </Button>
                  )}
            </form>
          </div>
          <a
            href="https://www.getprobo.com/"
            className="flex gap-1 text-sm font-medium text-txt-tertiary items-center py-6"
          >
            Powered by
            {" "}
            <Logo withPicto className="h-6" />
          </a>
        </div>
        {isDesktop && (
          <div className="bg-subtle h-full border-l border-border-solid min-h-0">
            {nda?.fileUrl && <PDFPreview src={nda.fileUrl} name={nda.fileName ?? ""} />}
          </div>
        )}
      </div>
    </div>
  );
}
