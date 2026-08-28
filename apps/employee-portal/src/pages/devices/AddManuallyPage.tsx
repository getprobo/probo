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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { ErrorState } from "@probo/ui/src/v2/ErrorState/ErrorState";
import { Spinner } from "@probo/ui/src/v2/Spinner/Spinner";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useParams } from "react-router";

import type { AddManuallyPageQuery } from "#/__generated__/core/AddManuallyPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { PageHeader } from "#/pages/_components/PageHeader";
import { EnrollmentInstructions } from "#/pages/devices/_components/EnrollmentInstructions";
import { addManuallyPage } from "#/pages/devices/_components/variants";
import { useEnrollDeviceManually } from "#/pages/devices/_lib/useEnrollDeviceManually";

export const addManuallyPageQuery = graphql`
  query AddManuallyPageQuery($organizationId: ID!) @throwOnFieldError {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canEnrollDevice: permission(action: "itam:device:enroll")
      }
    }
  }
`;

interface AddManuallyPageProps {
  queryRef: PreloadedQuery<AddManuallyPageQuery>;
}

export function AddManuallyPage({ queryRef }: AddManuallyPageProps) {
  const { t } = useTranslation("devices");
  const { t: tApp } = useTranslation();
  const { organizationId } = useParams();
  const slots = addManuallyPage();
  const { organization } = usePreloadedQuery<AddManuallyPageQuery>(
    addManuallyPageQuery,
    queryRef,
  );
  const { start, retry, isCreating, enrollment, failed } = useEnrollDeviceManually();
  const canEnroll = organization?.__typename === "Organization"
    && organization.canEnrollDevice;

  useEffect(() => {
    if (!canEnroll) {
      return;
    }
    start();
  }, [canEnroll, start]);

  if (organizationId === undefined) {
    throw new NotFoundError("organizationId is required");
  }

  if (organization?.__typename !== "Organization") {
    throw new NotFoundError("invalid type for organization node");
  }

  const header = (
    <PageHeader
      homeLabel={tApp("homePage.breadcrumb")}
      parent={{
        label: t("breadcrumb"),
        to: `/${organizationId}/devices`,
      }}
      currentLabel={t("addManuallyBreadcrumb")}
      title={t("addManually.title")}
    />
  );
  const devicesTo = `/${organizationId}/devices`;

  if (!canEnroll) {
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

  if (failed && enrollment === null) {
    return (
      <main className={slots.main()}>
        {header}
        <ErrorState
          title={tApp("common.error")}
          description={t("addManually.failed")}
          actions={(
            <div className={slots.errorActions()}>
              <Button
                size={3}
                variant="solid"
                color="neutral"
                highContrast
                loading={isCreating}
                onClick={retry}
              >
                {t("addManually.retry")}
              </Button>
              <ButtonLink
                to={devicesTo}
                size={3}
                variant="outline"
                color="neutral"
              >
                {t("addManually.back")}
              </ButtonLink>
            </div>
          )}
        />
      </main>
    );
  }

  return (
    <main className={slots.main()}>
      {header}
      {enrollment === null
        ? (
            <div className={slots.creating()}>
              <Spinner size={2} aria-label={t("addManually.creating")} />
              <Text size={2} color="neutral">
                {t("addManually.creating")}
              </Text>
            </div>
          )
        : (
            <>
              <EnrollmentInstructions
                enrollmentToken={enrollment.enrollmentToken}
                serverUrl={enrollment.serverUrl}
              />
              <ButtonLink
                to={devicesTo}
                size={3}
                variant="solid"
                color="neutral"
                highContrast
              >
                {t("addManually.done")}
              </ButtonLink>
            </>
          )}
    </main>
  );
}
