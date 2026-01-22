import { useCopy } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
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
  const { __ } = useTranslate();

  const confirm = useConfirm();
  const [isCopied, copy] = useCopy();

  const {
    samlConfigurations: { edges: samlConfigurations },
  } = useFragment<SAMLConfigurationListFragment$key>(fragment, fKey);

  const [deleteSAMLConfiguration]
    = useMutationWithToasts<SAMLConfigurationList_deleteMutation>(
      deleteMutation,
      {
        successMessage: "SAML configuration deleted successfully.",
        errorMessage: "Failed to delete SAML configuration. Please try again.",
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
        title: __("Delete SAML Configuration"),
        message: __(
          "Are you sure you want to delete the SAML configuration for "
          + config.emailDomain
          + "? This action cannot be undone.",
        ),
        label: __("Delete"),
        variant: "danger",
      },
    );
  };

  if (samlConfigurations.length === 0) {
    return (
      <Card padded>
        <div className="text-center py-12">
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            {__("No SAML Configurations")}
          </h3>
          <p className="text-gray-600 mb-6">
            {__(
              "Set up SAML 2.0 single sign-on for your organization by adding a configuration for each email domain.",
            )}
          </p>
        </div>
      </Card>
    );
  }

  return (
    <Table>
      <Thead>
        <Tr>
          <Th>{__("Email Domain")}</Th>
          <Th>{__("Domain Status")}</Th>
          <Th>{__("SAML Status")}</Th>
          <Th>{__("Enforcement")}</Th>
          <Th>{__("SSO URL")}</Th>
          <Th></Th>
        </Tr>
      </Thead>
      <Tbody>
        {samlConfigurations.map(({ node: config }) => (
          <Tr key={config.id}>
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
                  ? __("Verified")
                  : __("Pending Verification")}
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
                  ? __("Enabled")
                  : __("Disabled")}
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
                      {isCopied ? __("Copied!") : __("Copy URL")}
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
                            {__("Edit")}
                          </Button>
                        )}
                        {config.canDelete && (
                          <Button
                            variant="danger"
                            onClick={() => handleDelete(config)}
                          >
                            {__("Delete")}
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
                            {__("Verify Domain")}
                          </Button>
                        )}
                        {config.canDelete && (
                          <Button
                            variant="danger"
                            onClick={() => handleDelete(config)}
                          >
                            {__("Delete")}
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
