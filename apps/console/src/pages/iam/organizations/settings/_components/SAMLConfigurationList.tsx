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

import { useCopy } from "@probo/hooks";
import {
  Button,
  Card,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
} from "@probo/ui";
import { useCallback, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { SAMLConfigurationList_deleteMutation } from "#/__generated__/iam/SAMLConfigurationList_deleteMutation.graphql";
import type {
  SAMLConfigurationListFragment$data,
  SAMLConfigurationListFragment$key,
} from "#/__generated__/iam/SAMLConfigurationListFragment.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

const fragment = graphql`
  fragment SAMLConfigurationListFragment on Organization {
    samlConfigurations(first: 1000)
      @required(action: THROW)
      @connection(key: "SAMLConfigurationListFragment_samlConfigurations") {
      edges @required(action: THROW) {
        node {
          id
          emailDomain
          enforcementPolicy
          domainVerificationToken
          domainVerifiedAt
          testLoginUrl
          canUpdate: permission(action: "iam:saml-configuration:update")
          canDelete: permission(action: "iam:saml-configuration:delete")
        }
      }
    }
  }
`;

const deleteMutation = graphql`
  mutation SAMLConfigurationList_deleteMutation(
    $input: DeleteSAMLConfigurationInput!
    $connections: [ID!]!
  ) {
    deleteSAMLConfiguration(input: $input) {
      deletedSamlConfigurationId @deleteEdge(connections: $connections)
    }
  }
`;

export function SAMLConfigurationList(props: {
  fKey: SAMLConfigurationListFragment$key;
  onEdit: (id: string) => void;
  onVerifyDomain: (dnsVerificationToken: string) => void;
}) {
  const { fKey, onEdit, onVerifyDomain } = props;

  const organizationId = useOrganizationId();
  const { t } = useTranslation();

  const confirm = useConfirm();
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const copiedIdTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const copyId = useCallback((id: string) => {
    void navigator.clipboard.writeText(id);
    setCopiedId(id);
    clearTimeout(copiedIdTimer.current);
    copiedIdTimer.current = setTimeout(() => setCopiedId(null), 2000);
  }, []);
  const [isCopied, copy] = useCopy();

  const {
    samlConfigurations: { edges: samlConfigurations },
  } = useFragment<SAMLConfigurationListFragment$key>(fragment, fKey);

  const [deleteSAMLConfiguration]
    = useMutationWithToasts<SAMLConfigurationList_deleteMutation>(
      deleteMutation,
      {
        successMessage: t("samlConfigurationList.messages.deleted"),
        errorMessage: t("samlConfigurationList.errors.delete"),
      },
    );

  const handleDelete = (
    config: NodeOf<SAMLConfigurationListFragment$data["samlConfigurations"]>,
  ) => {
    confirm(
      async () => {
        await deleteSAMLConfiguration({
          variables: {
            input: {
              organizationId,
              samlConfigurationId: config.id,
            },
            connections: [
              ConnectionHandler.getConnectionID(
                organizationId,
                "SAMLConfigurationListFragment_samlConfigurations",
              ),
            ],
          },
        });
      },
      {
        title: t("samlConfigurationList.delete.title"),
        message: t("samlConfigurationList.delete.description", { domain: config.emailDomain }),
        label: t("samlConfigurationList.actions.delete"),
        variant: "danger",
      },
    );
  };

  if (samlConfigurations.length === 0) {
    return (
      <Card padded>
        <div className="text-center py-12">
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            {t("samlConfigurationList.empty.title")}
          </h3>
          <p className="text-gray-600 mb-6">
            {t("samlConfigurationList.empty.description")}
          </p>
        </div>
      </Card>
    );
  }

  return (
    <Table>
      <Thead>
        <Tr>
          <Th>{t("samlConfigurationList.columns.configurationId")}</Th>
          <Th>{t("samlConfigurationList.columns.emailDomain")}</Th>
          <Th>{t("samlConfigurationList.columns.domainStatus")}</Th>
          <Th>{t("samlConfigurationList.columns.samlStatus")}</Th>
          <Th>{t("samlConfigurationList.columns.enforcement")}</Th>
          <Th>{t("samlConfigurationList.columns.ssoUrl")}</Th>
          <Th></Th>
        </Tr>
      </Thead>
      <Tbody>
        {samlConfigurations.map(({ node: config }) => (
          <Tr key={config.id}>
            <Td>
              <button
                onClick={() => copyId(config.id)}
                className="font-mono text-xs text-gray-600 hover:text-gray-900"
                title={t("samlConfigurationList.actions.clickToCopy")}
              >
                {copiedId === config.id ? t("samlConfigurationList.actions.copied") : config.id}
              </button>
            </Td>
            <Td>
              <button
                onClick={() => onEdit(config.id)}
                className="font-semibold text-blue-600 hover:text-blue-800"
              >
                {config.emailDomain}
              </button>
            </Td>
            <Td>
              <span
                className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                  config.domainVerifiedAt
                    ? "bg-green-100 text-green-800"
                    : "bg-yellow-100 text-yellow-800"
                }`}
              >
                {config.domainVerifiedAt
                  ? t("samlConfigurationList.status.verified")
                  : t("samlConfigurationList.status.pendingVerification")}
              </span>
            </Td>
            <Td>
              <span
                className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                  config.enforcementPolicy !== "OFF"
                    ? "bg-green-100 text-green-800"
                    : "bg-gray-100 text-gray-800"
                }`}
              >
                {config.enforcementPolicy !== "OFF"
                  ? t("samlConfigurationList.status.enabled")
                  : t("samlConfigurationList.status.disabled")}
              </span>
            </Td>
            <Td>{config.enforcementPolicy}</Td>
            <Td>
              {config.domainVerifiedAt && config.enforcementPolicy !== "OFF"
                ? (
                    <button
                      onClick={() => copy(config.testLoginUrl)}
                      className="text-blue-600 hover:text-blue-800"
                    >
                      {isCopied ? t("samlConfigurationList.actions.copied") : t("samlConfigurationList.actions.copyUrl")}
                    </button>
                  )
                : (
                    <span className="text-gray-400">—</span>
                  )}
            </Td>
            <Td width={180} className="text-end">
              <div className="flex gap-2 justify-end">
                {config.domainVerifiedAt
                  ? (
                      <>
                        {config.canUpdate && (
                          <Button
                            variant="secondary"
                            onClick={() => onEdit(config.id)}
                          >
                            {t("samlConfigurationList.actions.edit")}
                          </Button>
                        )}
                        {config.canDelete && (
                          <Button
                            variant="danger"
                            onClick={() => handleDelete(config)}
                          >
                            {t("samlConfigurationList.actions.delete")}
                          </Button>
                        )}
                      </>
                    )
                  : (
                      <>
                        {config.canUpdate && !!config.domainVerificationToken && (
                          <Button
                            variant="primary"
                            onClick={() =>
                              onVerifyDomain(config.domainVerificationToken!)}
                          >
                            {t("samlConfigurationList.actions.verifyDomain")}
                          </Button>
                        )}
                        {config.canDelete && (
                          <Button
                            variant="danger"
                            onClick={() => handleDelete(config)}
                          >
                            {t("samlConfigurationList.actions.delete")}
                          </Button>
                        )}
                      </>
                    )}
              </div>
            </Td>
          </Tr>
        ))}
      </Tbody>
    </Table>
  );
}
