// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { useTranslate } from "@probo/i18n";
import { Breadcrumb, Button, Dialog, useDialogRef } from "@probo/ui";
import { Suspense, useState } from "react";
import {
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";

import type { EditSAMLConfigurationFormQuery } from "#/__generated__/iam/EditSAMLConfigurationFormQuery.graphql";
import type { SAMLSettingsPageQuery } from "#/__generated__/iam/SAMLSettingsPageQuery.graphql";

import {
  EditSAMLConfigurationForm,
  samlConfigurationFormQuery,
} from "./_components/EditSAMLConfigurationForm";
import { NewSAMLConfigurationForm } from "./_components/NewSAMLConfigurationForm";
import { SAMLConfigurationList } from "./_components/SAMLConfigurationList";
import { SAMLDomainVerifyDialog } from "./_components/SAMLDomainVerifyDialog";

export const samlSettingsPageQuery = graphql`
  query SAMLSettingsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canCreateSAMLConfiguration: permission(
          action: "iam:saml-configuration:create"
        )
        ...SAMLConfigurationListFragment
      }
    }
  }
`;

export function SAMLSettingsPage(props: {
  queryRef: PreloadedQuery<SAMLSettingsPageQuery>;
}) {
  const { queryRef } = props;

  const formDialogRef = useDialogRef();
  const domainDialogRef = useDialogRef();
  const [isEditing, setIsEditing] = useState<boolean>();
  const [domainVerificationToken, setDomainVerificationToken]
    = useState<string>();

  const { __ } = useTranslate();

  const { organization } = usePreloadedQuery<SAMLSettingsPageQuery>(samlSettingsPageQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("invalid node type");
  }
  const [formQueryRef, loadFormQuery]
    = useQueryLoader<EditSAMLConfigurationFormQuery>(samlConfigurationFormQuery);

  const handleOpenFormDialog = (samlConfigurationId?: string) => {
    setIsEditing(!!samlConfigurationId);
    if (samlConfigurationId) {
      loadFormQuery({ samlConfigurationId }, { fetchPolicy: "network-only" });
    }
    formDialogRef.current?.open();
  };
  const handleCloseFormDialog = () => {
    setIsEditing(false);
    formDialogRef.current?.close();
  };

  const handleOpenVerifyDomainDialog = (domainVerificationToken: string) => {
    setDomainVerificationToken(domainVerificationToken);
    domainDialogRef.current?.open();
  };
  const handleCloseVerifyDomainDialog = () => {
    setDomainVerificationToken("");
    formDialogRef.current?.close();
  };

  return (
    <>
      <div className="space-y-4">
        <div className="flex justify-between items-center">
          <h2 className="text-base font-medium">{__("SAML Single Sign-On")}</h2>
          {organization.canCreateSAMLConfiguration && (
            <Button onClick={() => handleOpenFormDialog()}>
              {__("Add Configuration")}
            </Button>
          )}
        </div>

        <SAMLConfigurationList
          fKey={organization}
          onEdit={(id: string) => handleOpenFormDialog(id)}
          onVerifyDomain={handleOpenVerifyDomainDialog}
        />
      </div>

      <Dialog
        ref={formDialogRef}
        onClose={handleCloseFormDialog}
        title={<Breadcrumb items={[__("SAML Settings"), __("Configure")]} />}
      >
        {isEditing
          ? (
              <Suspense>
                {formQueryRef && (
                  <EditSAMLConfigurationForm
                    queryRef={formQueryRef}
                    onUpdate={handleCloseFormDialog}
                  />
                )}
              </Suspense>
            )
          : (
              <NewSAMLConfigurationForm onCreate={handleCloseFormDialog} />
            )}
      </Dialog>

      <Dialog
        ref={domainDialogRef}
        onClose={handleCloseVerifyDomainDialog}
        title={
          <Breadcrumb items={[__("SAML Settings"), __("Verify Domain")]} />
        }
      >
        {domainVerificationToken && (
          <SAMLDomainVerifyDialog
            key={domainVerificationToken}
            domainVerificationToken={domainVerificationToken}
          />
        )}
      </Dialog>
    </>
  );
}
