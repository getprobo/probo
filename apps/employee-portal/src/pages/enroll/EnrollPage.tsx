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
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useSearchParams } from "react-router";

import type { EnrollPageQuery } from "#/__generated__/iam/EnrollPageQuery.graphql";
import { RelayProvider } from "#/lib/relay/RelayProvider";
import { DownloadStep } from "#/pages/devices/_components/DownloadStep";
import { OpenAgentStep } from "#/pages/devices/_components/OpenAgentStep";
import { ProgressStep } from "#/pages/devices/_components/ProgressStep";
import { ReviewStep } from "#/pages/devices/_components/ReviewStep";
import { registerDevicePage } from "#/pages/devices/_components/variants";
import { useEnrollDevice } from "#/pages/devices/_lib/useEnrollDevice";
import { TopBarUserMenu } from "#/pages/iam/_components/TopBar/TopBarUserMenu";
import { topBar } from "#/pages/iam/_components/TopBar/variants";

import { OrganizationStep } from "./_components/OrganizationStep";
import {
  ENROLL_STEPS,
  type EnrollStep,
  enrollStepIndex,
  maxEnrollStep,
  parseEnrollStep,
} from "./_lib/enrollSteps";

export const enrollPageQuery = graphql`
  query EnrollPageQuery @throwOnFieldError {
    viewer @required(action: THROW) {
      ...TopBarUserMenu_identity
      profiles(
        first: 1000
        orderBy: { direction: ASC, field: ORGANIZATION_NAME }
        filter: { states: [ACTIVE] }
      )
        @connection(key: "EnrollPage_profiles")
        @required(action: THROW) {
        edges @required(action: THROW) {
          node {
            organization @required(action: THROW) {
              id
              canEnrollDevice: permission(action: "itam:device:enroll")
            }
            ...OrganizationStep_profile
          }
        }
      }
    }
  }
`;

interface EnrollPageProps {
  queryRef: PreloadedQuery<EnrollPageQuery>;
}

export function EnrollPage({ queryRef }: EnrollPageProps) {
  const { t } = useTranslation("enroll");
  const { t: tDevices } = useTranslation("devices");
  const { t: tApp } = useTranslation();
  const [searchParams, setSearchParams] = useSearchParams();
  const slots = registerDevicePage();
  const bar = topBar();
  const tagline = tApp("topBar.tagline");
  const { viewer } = usePreloadedQuery<EnrollPageQuery>(enrollPageQuery, queryRef);
  const profiles = viewer.profiles.edges
    .map(({ node }) => node)
    .filter(profile => profile.organization.canEnrollDevice);
  const organizationIds = useMemo(
    () => new Set(profiles.map(profile => profile.organization.id)),
    [profiles],
  );
  const requestedOrganizationId = searchParams.get("organizationId");
  const selectedOrganizationId
    = requestedOrganizationId !== null && organizationIds.has(requestedOrganizationId)
      ? requestedOrganizationId
      : null;
  const requested = parseEnrollStep(searchParams.get("step"));
  const [reached, setReached] = useState(
    selectedOrganizationId === null ? "organization" : requested,
  );
  const clampedRequested = selectedOrganizationId === null ? "organization" : requested;
  const step = enrollStepIndex(clampedRequested) <= enrollStepIndex(reached)
    ? clampedRequested
    : reached;

  useEffect(() => {
    if (requested === step && (selectedOrganizationId === null || requestedOrganizationId === selectedOrganizationId)) {
      return;
    }

    setSearchParams((current) => {
      const params = new URLSearchParams(current);
      if (step === "organization") {
        params.delete("step");
      } else {
        params.set("step", step);
      }
      if (selectedOrganizationId === null) {
        params.delete("organizationId");
      } else {
        params.set("organizationId", selectedOrganizationId);
      }
      return params;
    }, { replace: true });
  }, [
    requested,
    requestedOrganizationId,
    selectedOrganizationId,
    setSearchParams,
    step,
  ]);

  function goToStep(next: EnrollStep) {
    if (enrollStepIndex(next) > enrollStepIndex(reached)) {
      return;
    }

    if (next !== "organization" && selectedOrganizationId === null) {
      return;
    }

    setSearchParams((current) => {
      const params = new URLSearchParams(current);
      if (next === "organization") {
        params.delete("step");
      } else {
        params.set("step", next);
      }
      return params;
    }, { replace: true });
  }

  function advanceTo(next: EnrollStep) {
    setReached(current => maxEnrollStep(current, next));
    setSearchParams((current) => {
      const params = new URLSearchParams(current);
      params.set("step", next);
      if (selectedOrganizationId !== null) {
        params.set("organizationId", selectedOrganizationId);
      }
      return params;
    }, { replace: true });
  }

  function handleOrganizationChange(organizationId: string) {
    if (organizationId !== selectedOrganizationId) {
      setReached("organization");
    }

    setSearchParams((current) => {
      const params = new URLSearchParams(current);
      params.set("organizationId", organizationId);
      params.delete("step");
      return params;
    }, { replace: true });
  }

  return (
    <div className="flex min-h-dvh flex-col bg-sand-2">
      <header className={bar.bar()}>
        <div className={bar.inner()}>
          <span className={bar.brand()}>
            <span className={bar.brandText()}>
              <Text
                size={2}
                weight="medium"
                color="neutral"
                highContrast
                className={bar.brandName()}
              >
                {tagline}
              </Text>
            </span>
          </span>
          <TopBarUserMenu identityKey={viewer} />
        </div>
      </header>

      <main className={slots.main()}>
        <Heading level={1} size={7} weight="medium" highContrast>
          {t("title")}
        </Heading>
        {profiles.length === 0
          ? (
              <ErrorState
                title={t("unavailable.title")}
                description={t("unavailable.description")}
                actions={(
                  <ButtonLink to="/" size={3} variant="solid" color="neutral" highContrast>
                    {t("unavailable.home")}
                  </ButtonLink>
                )}
              />
            )
          : (
              <div className={slots.body()}>
                <ol className={slots.stepper()}>
                  {ENROLL_STEPS.map((key, index) => {
                    const state = enrollStepIndex(key) < enrollStepIndex(step)
                      ? "complete"
                      : key === step
                        ? "current"
                        : "upcoming";
                    const title = key === "organization"
                      ? t("steps.organization.title")
                      : tDevices(`steps.${key}.title`);
                    const description = key === "organization"
                      ? t("steps.organization.description")
                      : tDevices(`steps.${key}.description`);

                    return (
                      <li key={key}>
                        <ProgressStep
                          number={String(index + 1)}
                          title={title}
                          description={description}
                          state={state}
                          onSelect={state === "complete" ? () => goToStep(key) : undefined}
                        />
                      </li>
                    );
                  })}
                </ol>
                <div className={slots.stage()}>
                  {step === "organization" && (
                    <OrganizationStep
                      profileKeys={profiles}
                      selectedOrganizationId={selectedOrganizationId}
                      onChange={handleOrganizationChange}
                      onContinue={() => advanceTo("review")}
                    />
                  )}
                  {step === "review" && (
                    <ReviewStep onContinue={() => advanceTo("download")} />
                  )}
                  {step === "download" && (
                    <DownloadStep onContinue={() => advanceTo("enroll")} />
                  )}
                  {step === "enroll" && selectedOrganizationId !== null && (
                    <RelayProvider>
                      <EnrollOpenAgent
                        key={selectedOrganizationId}
                        organizationId={selectedOrganizationId}
                      />
                    </RelayProvider>
                  )}
                </div>
              </div>
            )}
      </main>
    </div>
  );
}

function EnrollOpenAgent({ organizationId }: { organizationId: string }) {
  const enrollment = useEnrollDevice(organizationId);

  return (
    <OpenAgentStep
      enrollment={enrollment}
      organizationId={organizationId}
    />
  );
}
