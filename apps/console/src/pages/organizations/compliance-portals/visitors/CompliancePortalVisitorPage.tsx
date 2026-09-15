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

import { CaretLeftIcon, UserMinusIcon, UserPlusIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalVisitorPageActivateMutation } from "#/__generated__/core/CompliancePortalVisitorPageActivateMutation.graphql";
import type { CompliancePortalVisitorPageQuery } from "#/__generated__/core/CompliancePortalVisitorPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { useMutation } from "#/lib/relay/useMutation";

import { CompliancePortalDocumentAccessList } from "./_components/CompliancePortalDocumentAccessList";
import { CompliancePortalVisitorProfileCard } from "./_components/CompliancePortalVisitorProfileCard";
import { DeactivateVisitorDialog } from "./_components/DeactivateVisitorDialog";
import { ElectronicSignatureSection } from "./_components/ElectronicSignatureSection";
import { visitorDisplayName } from "./_lib/visitorIdentity";
import { visitorPage } from "./variants";

export const compliancePortalVisitorPageQuery = graphql`
  query CompliancePortalVisitorPageQuery(
    $accessId: ID!
    $filter: CompliancePortalAccessResourceFilter
  ) {
    access: node(id: $accessId) {
      __typename
      ... on CompliancePortalAccess {
        id
        state
        canGet: permission(action: "compliance-portal:portal-access:get")
        canUpdate: permission(action: "compliance-portal:portal-access:update")
        profile {
          fullName
          emailAddress
        }
        ndaSignature {
          ...ElectronicSignatureSectionFragment
        }
        ...CompliancePortalVisitorProfileCard_access
        ...CompliancePortalDocumentAccessList_access @arguments(filter: $filter)
      }
    }
  }
`;

const activateAccessMutation = graphql`
  mutation CompliancePortalVisitorPageActivateMutation($input: ActivateCompliancePortalAccessInput!) {
    activateCompliancePortalAccess(input: $input) {
      compliancePortalAccess {
        id
        state
      }
    }
  }
`;

interface CompliancePortalVisitorPageProps {
  queryRef: PreloadedQuery<CompliancePortalVisitorPageQuery>;
}

export function CompliancePortalVisitorPage({ queryRef }: CompliancePortalVisitorPageProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const { root, back, hero, callout } = visitorPage();
  const data = usePreloadedQuery<CompliancePortalVisitorPageQuery>(
    compliancePortalVisitorPageQuery,
    queryRef,
  );
  const [activateAccess, isActivating] = useMutation<CompliancePortalVisitorPageActivateMutation>(
    activateAccessMutation,
    {
      successMessage: t("visitorPage.messages.activated"),
      errorToast: t("visitorPage.errors.activate"),
    },
  );
  if (data.access?.__typename !== "CompliancePortalAccess" || !data.access.canGet) {
    throw new NotFoundError("Visitor not found");
  }

  const access = data.access;
  const canUpdate = access.canUpdate;
  const displayName = visitorDisplayName(
    access.profile.fullName,
    access.profile.emailAddress,
  );
  usePageTitle(displayName);

  function handleActivate() {
    void activateAccess({
      variables: { input: { id: access.id } },
    }).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  return (
    <div className={root()}>
      <Link to=".." size={2} color="neutral" underline={false} iconStart={<CaretLeftIcon />} className={back()}>
        {t("visitorPage.back")}
      </Link>
      {access.state === "DEACTIVATED" && (
        <Callout color="amber" className={callout()}>
          {t("visitorPage.deactivatedCallout")}
        </Callout>
      )}
      <div className={hero()}>
        <CompliancePortalVisitorProfileCard accessKey={access}>
          {canUpdate && (
            <>
              {access.state === "ACTIVE" && (
                <DeactivateVisitorDialog accessId={access.id}>
                  <Button
                    variant="soft"
                    color="red"
                    iconStart={<UserMinusIcon />}
                  >
                    {t("visitorPage.actions.deactivate")}
                  </Button>
                </DeactivateVisitorDialog>
              )}
              {access.state === "DEACTIVATED" && (
                <Button
                  variant="solid"
                  color="green"
                  iconStart={<UserPlusIcon />}
                  loading={isActivating}
                  onClick={handleActivate}
                >
                  {t("visitorPage.actions.activate")}
                </Button>
              )}
            </>
          )}
        </CompliancePortalVisitorProfileCard>
        {access.ndaSignature != null && (
          <ElectronicSignatureSection fragmentRef={access.ndaSignature} />
        )}
      </div>
      <CompliancePortalDocumentAccessList
        accessKey={access}
        accessId={access.id}
        canUpdate={canUpdate}
      />
    </div>
  );
}
