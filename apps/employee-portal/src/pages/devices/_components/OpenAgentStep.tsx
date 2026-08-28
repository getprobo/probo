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

import { AppWindowIcon, CheckCircleIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Spinner } from "@probo/ui/src/v2/Spinner/Spinner";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useParams } from "react-router";

import { NotFoundError } from "#/lib/relay/errors";
import { useEnrollDevice } from "#/pages/devices/_lib/useEnrollDevice";

import { RegisterDeviceCard } from "./RegisterDeviceCard";

export function OpenAgentStep() {
  const { t } = useTranslation("devices");
  const { organizationId } = useParams();
  const {
    openAgent,
    isCreating,
    isWaiting,
    isComplete,
    hasTimedOut,
    hostname,
  } = useEnrollDevice();

  if (organizationId == null) {
    throw new NotFoundError("organizationId is required");
  }

  if (isComplete) {
    return (
      <RegisterDeviceCard
        icon={<CheckCircleIcon />}
        title={hostname == null
          ? t("enroll.enrolled")
          : t("enroll.enrolledWithHostname", { hostname })}
        description={t("enroll.description")}
        action={(
          <ButtonLink
            to={`/${organizationId}`}
            size={3}
            variant="solid"
            color="neutral"
            highContrast
          >
            {t("enroll.home")}
          </ButtonLink>
        )}
      />
    );
  }

  if (isWaiting) {
    return (
      <RegisterDeviceCard
        icon={<AppWindowIcon />}
        title={t("enroll.title")}
        description={t("enroll.description")}
      >
        <div className="flex items-center gap-2">
          <Spinner size={2} aria-label={t("enroll.waiting")} />
          <Text size={2} color="neutral">
            {t("enroll.waiting")}
          </Text>
        </div>
      </RegisterDeviceCard>
    );
  }

  return (
    <RegisterDeviceCard
      icon={<AppWindowIcon />}
      title={t("enroll.title")}
      description={t("enroll.description")}
      action={(
        <Button
          size={3}
          variant="solid"
          color="neutral"
          highContrast
          loading={isCreating}
          onClick={() => {
            void openAgent();
          }}
        >
          {isCreating
            ? t("enroll.preparing")
            : hasTimedOut
              ? t("enroll.tryAgain")
              : t("enroll.open")}
        </Button>
      )}
    >
      {hasTimedOut && (
        <Text size={2} color="neutral">
          {t("enroll.timedOut")}
        </Text>
      )}
    </RegisterDeviceCard>
  );
}
