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

import { formatError } from "@probo/helpers";
import { Button, IconPlusLarge, Link, useToast } from "@probo/ui";
import { useEffect, useMemo, useRef } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useMutation } from "react-relay";

import type { AccessReviewOnboardingDoPanelFragment$key } from "#/__generated__/core/AccessReviewOnboardingDoPanelFragment.graphql";
import type { accessReviewSourceMutationsCreateMutation } from "#/__generated__/core/accessReviewSourceMutationsCreateMutation.graphql";
import type { AddAccessReviewSourceDialogConnectorProviderInfoFragment$key } from "#/__generated__/core/AddAccessReviewSourceDialogConnectorProviderInfoFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { AddAccessReviewSourceDialog, addAccessReviewSourceDialogConnectorProviderInfoFragment } from "#/pages/organizations/access-reviews/dialogs/AddAccessReviewSourceDialog";
import { createAccessReviewSourceMutation } from "#/pages/organizations/access-reviews/dialogs/accessReviewSourceMutations";
import { useSearchParams } from "react-router";

import {
  OnboardingStepActions,
  OnboardingStepDoCard,
} from "./OnboardingStepActions";

const accessReviewOnboardingDoPanelFragment = graphql`
  fragment AccessReviewOnboardingDoPanelFragment on Organization {
    canCreateSource: permission(action: "access-review:source:create")
    accessReviewSources(first: 50) {
      __id
      edges {
        node {
          id
          connectorId
          connector {
            provider
          }
        }
      }
    }
  }
`;

function clearOAuthCallbackParams(params: URLSearchParams) {
  params.delete("connector_id");
  params.delete("provider");
  params.delete("error");
  return params;
}

type Props = {
  accessReviewDrivers: AddAccessReviewSourceDialogConnectorProviderInfoFragment$key;
  onContinue: () => void;
  onDefer: () => void;
  organizationFKey: AccessReviewOnboardingDoPanelFragment$key;
};

export function AccessReviewOnboardingDoPanel(props: Props) {
  const { accessReviewDrivers, onContinue, onDefer, organizationFKey } = props;
  const { t } = useTranslation("organizations/onboarding");
  const { t: tSources } = useTranslation();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const [searchParams, setSearchParams] = useSearchParams();
  const processedConnectorIdRef = useRef<string | null>(null);

  const organization = useFragment<AccessReviewOnboardingDoPanelFragment$key>(
    accessReviewOnboardingDoPanelFragment,
    organizationFKey,
  );

  const connectorProviderInfos = useFragment<AddAccessReviewSourceDialogConnectorProviderInfoFragment$key>(
    addAccessReviewSourceDialogConnectorProviderInfoFragment,
    accessReviewDrivers,
  );

  const accessReviewSources = organization.accessReviewSources;
  const complete = accessReviewSources.edges.length > 0;

  const existingSourceProviders = useMemo(
    () =>
      accessReviewSources.edges
        .map(edge => edge.node.connector?.provider)
        .filter((p): p is NonNullable<typeof p> => p != null),
    [accessReviewSources.edges],
  );

  const [createAccessReviewSource, isCreatingSource]
    = useMutation<accessReviewSourceMutationsCreateMutation>(
      createAccessReviewSourceMutation,
    );

  const callbackConnectorId = searchParams.get("connector_id");
  const callbackProvider = searchParams.get("provider");
  const callbackError = searchParams.get("error");
  const hasSourceForCallback = !!callbackConnectorId
    && accessReviewSources.edges.some(
      edge => edge.node.connectorId === callbackConnectorId,
    );

  useEffect(() => {
    if (!callbackConnectorId) return;

    if (hasSourceForCallback) {
      const createInFlight
        = processedConnectorIdRef.current === callbackConnectorId;
      if (callbackError && !createInFlight) {
        toast({
          title: tSources("accessReviewSourcesTab.messages.error"),
          description: callbackError,
          variant: "error",
        });
      }
      if (!createInFlight) {
        processedConnectorIdRef.current = null;
        setSearchParams(clearOAuthCallbackParams, { replace: true });
      }
      return;
    }

    if (processedConnectorIdRef.current === callbackConnectorId || isCreatingSource) {
      return;
    }
    processedConnectorIdRef.current = callbackConnectorId;

    const providerInfo = callbackProvider
      ? connectorProviderInfos.find(p => p.provider === callbackProvider)
      : null;
    const sourceName = providerInfo?.displayName ?? callbackProvider ?? "Source";

    createAccessReviewSource({
      variables: {
        input: {
          organizationId,
          connectorId: callbackConnectorId,
          name: sourceName,
          csvData: null,
        },
        connections: [accessReviewSources.__id],
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          processedConnectorIdRef.current = null;
          setSearchParams(clearOAuthCallbackParams, { replace: true });
          toast({
            title: tSources("accessReviewSourcesTab.messages.error"),
            description: formatError(
              tSources("accessReviewSourcesTab.errors.create"),
              errors,
            ),
            variant: "error",
          });
          return;
        }
        if (callbackError) {
          toast({
            title: tSources("accessReviewSourcesTab.messages.error"),
            description: callbackError,
            variant: "error",
          });
        } else {
          toast({
            title: tSources("accessReviewSourcesTab.messages.success"),
            description: tSources("accessReviewSourcesTab.messages.created"),
            variant: "success",
          });
        }
        processedConnectorIdRef.current = null;
        setSearchParams(clearOAuthCallbackParams, { replace: true });
      },
      onError(error) {
        processedConnectorIdRef.current = null;
        setSearchParams(clearOAuthCallbackParams, { replace: true });
        toast({
          title: tSources("accessReviewSourcesTab.messages.error"),
          description: formatError(
            tSources("accessReviewSourcesTab.errors.create"),
            error,
          ),
          variant: "error",
        });
      },
    });
  }, [
    accessReviewSources.__id,
    accessReviewSources.edges,
    callbackConnectorId,
    callbackError,
    callbackProvider,
    connectorProviderInfos,
    createAccessReviewSource,
    hasSourceForCallback,
    isCreatingSource,
    organizationId,
    setSearchParams,
    tSources,
    toast,
  ]);

  const settingsHref = `/organizations/${organizationId}/access-reviews/sources`;

  return (
    <OnboardingStepDoCard
      title={t("steps.accessReview.doTitle")}
      description={t("steps.accessReview.doDescription")}
      complete={complete}
      actions={(
        <OnboardingStepActions
          continueDisabled={!complete}
          onContinue={onContinue}
          onDefer={onDefer}
          settingsLink={(
            <Link to={settingsHref} variant="secondary" size="2">
              {t("actions.openInSettings")}
            </Link>
          )}
        />
      )}
    >
      {organization.canCreateSource
        ? (
            <AddAccessReviewSourceDialog
              organizationId={organizationId}
              connectionId={accessReviewSources.__id}
              providerInfos={connectorProviderInfos}
              existingSourceProviders={existingSourceProviders}
            >
              <Button icon={IconPlusLarge}>
                {t("steps.accessReview.addConnector")}
              </Button>
            </AddAccessReviewSourceDialog>
          )
        : (
            <p className="text-2 text-txt-secondary">
              {t("steps.accessReview.noPermission")}
            </p>
          )}
    </OnboardingStepDoCard>
  );
}
