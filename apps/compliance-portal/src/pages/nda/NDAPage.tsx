// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import {
  CaretLeftIcon,
  CaretRightIcon,
  MagnifyingGlassMinusIcon,
  MagnifyingGlassPlusIcon,
} from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Separator } from "@probo/ui/src/v2/Separator/Separator";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { startTransition, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePreloadedQuery, useRefetchableFragment } from "react-relay";
import { Navigate, useSearchParams } from "react-router";

import { HeaderBand } from "#/components/HeaderBand/HeaderBand";
import { getSafeContinueUrl } from "#/lib/auth/continueUrl";
import { useMutation } from "#/lib/relay/useMutation";
import { PdfPreview, type PdfPreviewHandle } from "#/pages/documents/_components/PdfPreview";

import type { NDAPageAcceptMutation } from "./__generated__/NDAPageAcceptMutation.graphql";
import type { NDAPageFragment$key } from "./__generated__/NDAPageFragment.graphql";
import type { NDAPageQuery as NDAPageQueryType } from "./__generated__/NDAPageQuery.graphql";
import type { NDAPageRecordEventMutation } from "./__generated__/NDAPageRecordEventMutation.graphql";
import type { NDAPageRefetchQuery } from "./__generated__/NDAPageRefetchQuery.graphql";
import { ndaPage } from "./variants";

const POLL_INTERVAL_MS = 1500;
const MIN_SCALE = 0.5;
const MAX_SCALE = 3;

function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

export const ndaPageQuery = graphql`
  query NDAPageQuery {
    viewer {
      id
    }
    currentTrustCenter @required(action: THROW) {
      title
      nonDisclosureAgreement {
        fileUrl
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

const acceptSignatureMutation = graphql`
  mutation NDAPageAcceptMutation($input: AcceptElectronicSignatureInput!) {
    acceptElectronicSignature(input: $input) {
      signature {
        id
        status
      }
    }
  }
`;

const recordSigningEventMutation = graphql`
  mutation NDAPageRecordEventMutation($input: RecordSigningEventInput!) {
    recordSigningEvent(input: $input) {
      success
    }
  }
`;

interface NDAPageProps {
  queryRef: PreloadedQuery<NDAPageQueryType>;
}

// Non-Disclosure Agreement gate: the user reviews the NDA (rendered in the body)
// and signs it via the header action. Signing records the consent events, accepts
// the electronic signature, polls until it is sealed, then returns to the
// continue URL. Reached from the route boundary on NDA_SIGNATURE_REQUIRED.
export function NDAPage({ queryRef }: NDAPageProps) {
  const { t } = useTranslation("nda");
  const [searchParams] = useSearchParams();
  const documentViewedRef = useRef(false);
  const pdfRef = useRef<PdfPreviewHandle>(null);
  const [numPages, setNumPages] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [scale, setScale] = useState(1);

  const data = usePreloadedQuery<NDAPageQueryType>(ndaPageQuery, queryRef);
  const trustCenter = data.currentTrustCenter;
  const [fragment, refetch] = useRefetchableFragment<NDAPageRefetchQuery, NDAPageFragment$key>(
    ndaPageFragment,
    trustCenter,
  );

  const nda = trustCenter.nonDisclosureAgreement;
  const signature = fragment.nonDisclosureAgreement.viewerSignature;

  const safeContinueUrl = getSafeContinueUrl(searchParams.get("continue"));

  const [acceptSignature, isAccepting] = useMutation<NDAPageAcceptMutation>(
    acceptSignatureMutation,
    { errorToast: false },
  );
  const [recordSigningEvent] = useMutation<NDAPageRecordEventMutation>(
    recordSigningEventMutation,
    { errorToast: false },
  );

  const isProcessing = signature?.status === "ACCEPTED" || signature?.status === "PROCESSING";
  const isFailed = signature?.status === "FAILED";
  const isCompleted = signature?.status === "COMPLETED";

  // Once the signature is sealed, leave the gate and resume where the user was.
  useEffect(() => {
    if (isCompleted) {
      window.location.href = safeContinueUrl;
    }
  }, [isCompleted, safeContinueUrl]);

  // While the backend seals the signature, poll the fragment for the new status.
  useEffect(() => {
    if (!isProcessing) {
      return;
    }
    const interval = setInterval(() => {
      startTransition(() => {
        refetch({}, { fetchPolicy: "network-only" });
      });
    }, POLL_INTERVAL_MS);
    return () => clearInterval(interval);
  }, [isProcessing, refetch]);

  // Record that the document was viewed once, on first render of a pending gate.
  useEffect(() => {
    if (signature?.status === "PENDING" && !documentViewedRef.current) {
      documentViewedRef.current = true;
      void recordSigningEvent({
        variables: { input: { signatureId: signature.id, eventType: "DOCUMENT_VIEWED" } },
      }).catch(() => {});
    }
  }, [signature, recordSigningEvent]);

  const handleAccept = () => {
    if (!signature) {
      return;
    }

    if (signature.status === "PENDING") {
      void recordSigningEvent({
        variables: { input: { signatureId: signature.id, eventType: "FULL_NAME_TYPED" } },
      }).catch(() => {});
    }

    // Consent + acceptance are the critical steps: surface failures (via the
    // default mutation error toast) so the user can retry, instead of leaving
    // the sign button apparently inert. The fire-and-forget events above stay
    // silent (errorToast: false on the hook).
    void recordSigningEvent(
      {
        variables: { input: { signatureId: signature.id, eventType: "CONSENT_GIVEN" } },
        onCompleted: () => {
          void acceptSignature(
            { variables: { input: { signatureId: signature.id } } },
            { errorToast: true },
          ).catch(() => {});
        },
      },
      { errorToast: true },
    ).catch(() => {});
  };

  const movePage = (direction: 1 | -1) => {
    const next = clamp(currentPage + direction, 1, numPages);
    pdfRef.current?.scrollToPage(next);
    setCurrentPage(next);
  };

  if (!data.viewer) {
    return <Navigate to="/" replace />;
  }

  if (!nda || !signature) {
    return <Navigate to="/" replace />;
  }

  // Signature already sealed: the effect above redirects to the continue URL;
  // render nothing meanwhile so we don't flash the sign UI or the home page.
  if (isCompleted) {
    return null;
  }

  const slots = ndaPage();

  return (
    <div className={slots.root()}>
      <HeaderBand flushBottomSpace>
        <div className={slots.header()}>
          <div className={slots.text()}>
            <Heading level={1} size={7} weight="medium" highContrast>
              {t("title")}
            </Heading>
            <Text size={2} color="neutral">
              {t("subtitle", { name: trustCenter.title })}
            </Text>
            {signature.consentText != null && (
              <Text size={1} color="faint" className={slots.consent()}>
                {signature.consentText}
              </Text>
            )}
          </div>

          {isFailed && (
            <Callout color="red" variant="surface">
              {signature.lastError ?? t("failedDescription")}
            </Callout>
          )}

          <div className={slots.toolbar()}>
            <div className={slots.toolbarStart()}>
              {numPages > 0 && (
                <>
                  <div className={slots.controls()}>
                    <IconButton
                      variant="ghost"
                      color="neutral"
                      aria-label={t("common.previousPage")}
                      disabled={currentPage <= 1}
                      onClick={() => movePage(-1)}
                    >
                      <CaretLeftIcon />
                    </IconButton>
                    <Text size={2} color="neutral">
                      {t("common.pageOf", { current: currentPage, total: numPages })}
                    </Text>
                    <IconButton
                      variant="ghost"
                      color="neutral"
                      aria-label={t("common.nextPage")}
                      disabled={currentPage >= numPages}
                      onClick={() => movePage(1)}
                    >
                      <CaretRightIcon />
                    </IconButton>
                  </div>
                  <Separator orientation="vertical" className={slots.separator()} />
                  <div className={slots.controls()}>
                    <IconButton
                      variant="ghost"
                      color="neutral"
                      aria-label={t("common.zoomOut")}
                      onClick={() => setScale(value => clamp(value * 0.8, MIN_SCALE, MAX_SCALE))}
                    >
                      <MagnifyingGlassMinusIcon />
                    </IconButton>
                    <Text size={2} color="neutral">
                      {`${Math.round(scale * 100)}%`}
                    </Text>
                    <IconButton
                      variant="ghost"
                      color="neutral"
                      aria-label={t("common.zoomIn")}
                      onClick={() => setScale(value => clamp(value * 1.25, MIN_SCALE, MAX_SCALE))}
                    >
                      <MagnifyingGlassPlusIcon />
                    </IconButton>
                  </div>
                </>
              )}
            </div>
            <div className={slots.actions()}>
              <Button
                type="button"
                color="neutral"
                highContrast
                loading={isAccepting || isProcessing}
                disabled={isProcessing}
                onClick={handleAccept}
              >
                {isProcessing
                  ? t("sealing")
                  : isFailed
                    ? t("tryAgain")
                    : t("reviewAndSign")}
              </Button>
            </div>
          </div>
        </div>
      </HeaderBand>

      <div className={slots.body()}>
        {nda.fileUrl
          ? (
              <PdfPreview
                ref={pdfRef}
                file={nda.fileUrl}
                scale={scale}
                onNumPages={setNumPages}
                onVisiblePageChange={setCurrentPage}
              />
            )
          : <div className={slots.stage()} />}
      </div>
    </div>
  );
}
