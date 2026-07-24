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

import { LaptopIcon } from "@phosphor-icons/react";
import { sprintf } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, Card } from "@probo/ui";
import { useCallback, useMemo, useState } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Link } from "react-router";
import { graphql } from "relay-runtime";

import type { EnrollDevicePageQuery } from "#/__generated__/iam/EnrollDevicePageQuery.graphql";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import {
  EnrollDeviceButton,
  type EnrollmentSession,
} from "../../organizations/enroll/_components/EnrollDeviceButton";

import { EnrollOrganizationPicker } from "./_components/EnrollOrganizationPicker";

export const enrollDevicePageQuery = graphql`
  query EnrollDevicePageQuery {
    viewer @required(action: THROW) {
      profiles(
        first: 1000
        orderBy: { direction: ASC, field: ORGANIZATION_NAME }
        filter: { state: ACTIVE }
      ) @required(action: THROW) {
        edges @required(action: THROW) {
          node @required(action: THROW) {
            id
            organization @required(action: THROW) {
              id
              name
            }
          }
        }
      }
    }
  }
`;

interface EnrollDevicePageProps {
  queryRef: PreloadedQuery<EnrollDevicePageQuery>;
}

export function EnrollDevicePage({ queryRef }: EnrollDevicePageProps) {
  const { __ } = useTranslate();
  usePageTitle(__("Enroll device"));

  const { viewer } = usePreloadedQuery<EnrollDevicePageQuery>(
    enrollDevicePageQuery,
    queryRef,
  );

  const organizations = useMemo(
    () => viewer.profiles.edges.map(edge => edge.node.organization),
    [viewer.profiles.edges],
  );

  const [manualOrganizationId, setManualOrganizationId] = useState<string | null>(
    null,
  );
  const [step, setStep] = useState<"intro" | "organization" | "enroll">("intro");
  const [enrollmentSession, setEnrollmentSession] = useState<EnrollmentSession | null>(
    null,
  );
  const [isEnrollmentComplete, setIsEnrollmentComplete] = useState(false);
  const handleEnrollmentComplete = useCallback(() => {
    setIsEnrollmentComplete(true);
  }, []);
  const stepIndexByName = {
    intro: 1,
    organization: 2,
    enroll: 3,
  } as const;

  const selectedOrganizationId = useMemo(() => {
    if (
      manualOrganizationId
      && organizations.some(org => org.id === manualOrganizationId)
    ) {
      return manualOrganizationId;
    }

    return organizations[0]?.id ?? null;
  }, [manualOrganizationId, organizations]);

  const handleOrganizationChange = (organizationID: string) => {
    setManualOrganizationId(organizationID);
    setEnrollmentSession(current =>
      current !== null && current.organizationId !== organizationID ? null : current,
    );
  };

  const activeEnrollmentSession
    = enrollmentSession !== null
      && selectedOrganizationId !== null
      && enrollmentSession.organizationId === selectedOrganizationId
      ? enrollmentSession
      : null;

  const activeStepIndex = stepIndexByName[step];

  const steps = [
    {
      key: "intro",
      title: __("Privacy"),
      description: __("Review collected data"),
    },
    {
      key: "organization",
      title: __("Organization"),
      description: __("Choose destination workspace"),
    },
    {
      key: "enroll",
      title: __("Open and wait"),
      description: __("Finish setup in the desktop agent"),
    },
  ] as const;

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 py-8">
      {organizations.length === 0
        ? (
            <Card className="space-y-4 border-border-low p-6">
              <div className="space-y-1">
                <h2 className="text-lg font-medium">
                  {__("Enrollment unavailable")}
                </h2>
                <p className="text-sm text-txt-secondary">
                  {__(
                    "You do not have permission to enroll devices in any organization.",
                  )}
                </p>
              </div>
              <Button variant="secondary" asChild>
                <Link to="/">{__("Back to organizations")}</Link>
              </Button>
            </Card>
          )
        : (
            <Card className="overflow-hidden border-border-low p-0">
              <div className="grid md:grid-cols-[260px_minmax(0,1fr)]">
                <aside className="flex flex-col border-b border-border-low bg-subtle/30 p-6 md:min-h-[560px] md:border-b-0 md:border-r">
                  <div className="flex items-center gap-3">
                    <div className="flex size-9 items-center justify-center rounded-lg border border-border-low bg-level-1 text-txt-primary">
                      <LaptopIcon size={18} weight="duotone" />
                    </div>
                    <div>
                      <p className="text-sm font-semibold text-txt-primary">{__("Device enrollment")}</p>
                      <p className="text-xs text-txt-secondary">{__("Setup")}</p>
                    </div>
                  </div>

                  <div className="mt-8 space-y-2">
                    {steps.map((item, index) => {
                      const stepNumber = index + 1;
                      const isActive = stepNumber === activeStepIndex;
                      const isComplete = stepNumber < activeStepIndex;

                      return (
                        <div
                          key={item.key}
                          className={[
                            "rounded-lg px-3 py-2",
                            isActive ? "bg-level-1" : "",
                          ].join(" ")}
                        >
                          <div className="flex items-start gap-2.5">
                            <span
                              className={[
                                "mt-0.5 flex size-5 shrink-0 items-center justify-center rounded-full text-[10px] font-semibold",
                                isComplete || isActive
                                  ? "bg-primary text-invert"
                                  : "bg-level-1 text-txt-secondary",
                              ].join(" ")}
                            >
                              {stepNumber}
                            </span>
                            <div className="min-w-0">
                              <p className="text-sm font-medium text-txt-primary">{item.title}</p>
                              <p className="text-xs text-txt-secondary">{item.description}</p>
                            </div>
                          </div>
                        </div>
                      );
                    })}
                  </div>

                  <div className="mt-8 md:mt-auto">
                    <p className="text-xs font-semibold uppercase tracking-[0.12em] text-txt-secondary">
                      {sprintf(__("Step %s of %s"), activeStepIndex, steps.length)}
                    </p>
                    <div className="mt-3 flex gap-2">
                      {steps.map((item, index) => (
                        <span
                          key={item.key}
                          className={[
                            "h-1.5 flex-1 rounded-full",
                            index < activeStepIndex ? "bg-primary" : "bg-border-low",
                          ].join(" ")}
                        />
                      ))}
                    </div>
                  </div>
                </aside>

                <section className="space-y-6 p-6 md:p-8">
                  {step === "intro" && (
                    <section className="space-y-5">
                      <header className="space-y-2">
                        <h1 className="text-2xl font-semibold tracking-tight">{__("Before you start")}</h1>
                        <p className="text-sm leading-6 text-txt-secondary">
                          {__(
                            "Probo collects the following device metadata for inventory and posture reporting:",
                          )}
                        </p>
                      </header>
                      <ul className="list-disc space-y-1.5 pl-5 text-sm text-txt-secondary">
                        <li>{__("Device identity: hardware UUID, hostname, and serial number (when available).")}</li>
                        <li>{__("System details: platform, OS version, and Probo agent version.")}</li>
                        <li>{__("Activity signals: enrollment time, heartbeats, and posture check results.")}</li>
                      </ul>
                      <div className="flex flex-wrap gap-3 pt-2">
                        <Button onClick={() => setStep("organization")}>
                          {__("Continue")}
                        </Button>
                      </div>
                    </section>
                  )}

                  {step === "organization" && (
                    <section className="space-y-5">
                      <header className="space-y-2">
                        <h1 className="text-2xl font-semibold tracking-tight">{__("Choose organization")}</h1>
                        <p className="text-sm leading-6 text-txt-secondary">
                          {__(
                            "Pick which organization will own and manage this device.",
                          )}
                        </p>
                      </header>
                      <EnrollOrganizationPicker
                        organizations={organizations.map(organization => ({
                          id: organization.id,
                          name: organization.name,
                        }))}
                        selectedOrganizationId={selectedOrganizationId}
                        onChange={handleOrganizationChange}
                      />
                      <div className="flex flex-wrap gap-3">
                        <Button
                          onClick={() => setStep("enroll")}
                          disabled={selectedOrganizationId == null}
                        >
                          {__("Continue")}
                        </Button>
                        <Button variant="secondary" onClick={() => setStep("intro")}>
                          {__("Back")}
                        </Button>
                      </div>
                    </section>
                  )}

                  {step === "enroll" && (
                    <section className="space-y-5">
                      <header className="space-y-2">
                        <h1 className="text-2xl font-semibold tracking-tight">{__("Open the Probo agent")}</h1>
                        <p className="text-sm leading-6 text-txt-secondary">
                          {__(
                            "Open the desktop agent to finish setup, then keep this page open until enrollment is confirmed.",
                          )}
                        </p>
                      </header>
                      <CoreRelayProvider>
                        <EnrollDeviceButton
                          organizationId={selectedOrganizationId}
                          session={activeEnrollmentSession}
                          onSessionCreated={setEnrollmentSession}
                          onComplete={handleEnrollmentComplete}
                        />
                      </CoreRelayProvider>
                      {!isEnrollmentComplete && (
                        <div>
                          <Button variant="secondary" onClick={() => setStep("organization")}>
                            {__("Back")}
                          </Button>
                        </div>
                      )}
                    </section>
                  )}
                </section>
              </div>
            </Card>
          )}
    </div>
  );
}
