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

import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { ErrorState } from "@probo/ui/src/v2/ErrorState/ErrorState";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useParams, useSearchParams } from "react-router";

import type { RegisterDevicePageQuery } from "#/__generated__/core/RegisterDevicePageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { PageHeader } from "#/pages/_components/PageHeader";
import { DownloadStep } from "#/pages/devices/_components/DownloadStep";
import { OpenAgentStep } from "#/pages/devices/_components/OpenAgentStep";
import { ProgressStep } from "#/pages/devices/_components/ProgressStep";
import { ReviewStep } from "#/pages/devices/_components/ReviewStep";
import { registerDevicePage } from "#/pages/devices/_components/variants";
import {
  maxRegisterDeviceStep,
  parseRegisterDeviceStep,
  REGISTER_DEVICE_STEPS,
  type RegisterDeviceStep,
  registerDeviceStepIndex,
} from "#/pages/devices/_lib/registerDeviceSteps";

export const registerDevicePageQuery = graphql`
  query RegisterDevicePageQuery($organizationId: ID!) @throwOnFieldError {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canEnrollDevice: permission(action: "itam:device:enroll")
      }
    }
  }
`;

interface RegisterDevicePageProps {
  queryRef: PreloadedQuery<RegisterDevicePageQuery>;
}

export function RegisterDevicePage({ queryRef }: RegisterDevicePageProps) {
  const { t } = useTranslation("devices");
  const { t: tApp } = useTranslation();
  const { organizationId } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  const slots = registerDevicePage();
  const { organization } = usePreloadedQuery<RegisterDevicePageQuery>(
    registerDevicePageQuery,
    queryRef,
  );
  const requested = parseRegisterDeviceStep(searchParams.get("step"));
  const [reached, setReached] = useState(requested);
  const step = registerDeviceStepIndex(requested) <= registerDeviceStepIndex(reached)
    ? requested
    : reached;

  useEffect(() => {
    if (requested === step) {
      return;
    }

    setSearchParams((current) => {
      const params = new URLSearchParams(current);
      if (step === "review") {
        params.delete("step");
      } else {
        params.set("step", step);
      }
      return params;
    }, { replace: true });
  }, [requested, setSearchParams, step]);

  if (organizationId === undefined) {
    throw new NotFoundError("organizationId is required");
  }

  if (organization?.__typename !== "Organization") {
    throw new NotFoundError("invalid type for organization node");
  }

  function goToStep(next: RegisterDeviceStep) {
    if (registerDeviceStepIndex(next) > registerDeviceStepIndex(reached)) {
      return;
    }

    setSearchParams((current) => {
      const params = new URLSearchParams(current);
      if (next === "review") {
        params.delete("step");
      } else {
        params.set("step", next);
      }
      return params;
    }, { replace: true });
  }

  function advanceTo(next: RegisterDeviceStep) {
    setReached(current => maxRegisterDeviceStep(current, next));
    setSearchParams((current) => {
      const params = new URLSearchParams(current);
      params.set("step", next);
      return params;
    }, { replace: true });
  }

  const header = (
    <PageHeader
      homeLabel={tApp("homePage.breadcrumb")}
      parent={{
        label: t("breadcrumb"),
        to: `/${organizationId}/devices`,
      }}
      currentLabel={t("registerBreadcrumb")}
      title={t("title")}
    />
  );

  if (!organization.canEnrollDevice) {
    return (
      <main className={slots.main()}>
        {header}
        <ErrorState
          title={t("unavailable.title")}
          description={t("unavailable.description")}
          actions={(
            <ButtonLink
              to={`/${organizationId}`}
              size={3}
              variant="solid"
              color="neutral"
              highContrast
            >
              {t("unavailable.home")}
            </ButtonLink>
          )}
        />
      </main>
    );
  }

  return (
    <main className={slots.main()}>
      {header}
      <div className={slots.body()}>
        <ol className={slots.stepper()}>
          {REGISTER_DEVICE_STEPS.map((key, index) => {
            const state = registerDeviceStepIndex(key) < registerDeviceStepIndex(step)
              ? "complete"
              : key === step
                ? "current"
                : "upcoming";

            return (
              <li key={key}>
                <ProgressStep
                  number={String(index + 1)}
                  title={t(`steps.${key}.title`)}
                  description={t(`steps.${key}.description`)}
                  state={state}
                  onSelect={state === "complete" ? () => goToStep(key) : undefined}
                />
              </li>
            );
          })}
        </ol>
        <div className={slots.stage()}>
          {step === "review" && (
            <ReviewStep onContinue={() => advanceTo("download")} />
          )}
          {step === "download" && (
            <DownloadStep onContinue={() => advanceTo("enroll")} />
          )}
          {step === "enroll" && <OpenAgentStep />}
        </div>
      </div>
    </main>
  );
}
