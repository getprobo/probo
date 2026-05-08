// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import {
  awsWizardInitialState,
  canAdvanceAwsStep,
  formatError,
  type GraphQLError,
  reduceAwsWizard,
  safeOpenUrl,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Field,
  IconArrowLink,
  IconCircleInfo,
  useToast,
} from "@probo/ui";
import { useReducer } from "react";
import { graphql, useMutation } from "react-relay";

import type { AwsConnectWizardCreateMutation } from "#/__generated__/core/AwsConnectWizardCreateMutation.graphql";
import type { AwsConnectWizardGenerateAssetsMutation } from "#/__generated__/core/AwsConnectWizardGenerateAssetsMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const generateAssetsMutation = graphql`
  mutation AwsConnectWizardGenerateAssetsMutation(
    $input: GenerateCloudAccountInstallAssetsInput!
  ) {
    generateCloudAccountInstallAssets(input: $input) {
      assets {
        __typename
        ... on AWSInstallAssets {
          quickCreateURL
          externalId
          principalArn
          requiredActions
        }
      }
    }
  }
`;

const createMutation = graphql`
  mutation AwsConnectWizardCreateMutation(
    $input: CreateCloudAccountInput!
    $connections: [ID!]!
  ) {
    createCloudAccount(input: $input) {
      cloudAccount
        @prependNode(
          connections: $connections
          edgeTypeName: "CloudAccountEdge"
        ) {
        id
        provider
        label
        status
        scope {
          kind
          identifier
        }
        lastVerifiedAt
      }
      verifyStatus
      lastProbeError
    }
  }
`;

type Props = {
  connectionId: string;
  onComplete: () => void;
  onBack: () => void;
};

export function AwsConnectWizard(props: Props) {
  const { connectionId, onComplete, onBack } = props;
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const [state, dispatch] = useReducer(reduceAwsWizard, awsWizardInitialState);

  const [generateAssets, isGenerating]
    = useMutation<AwsConnectWizardGenerateAssetsMutation>(generateAssetsMutation);
  const [createCloudAccount]
    = useMutation<AwsConnectWizardCreateMutation>(createMutation);

  const handleGenerateAssets = () => {
    generateAssets({
      variables: {
        input: {
          organizationId,
          provider: "AWS",
          scopeKind: state.scopeKind,
          scopeIdentifier: state.scopeIdentifier,
          modules: ["ACCESS_REVIEW"],
          awsRegion: state.region,
        },
      },
      onCompleted(response, errors) {
        if (errors?.length) {
          dispatch({
            type: "install-assets-error",
            message: errors.map(e => e.message).join(", "),
          });
          return;
        }
        const assets = response.generateCloudAccountInstallAssets.assets;
        if (assets.__typename !== "AWSInstallAssets") {
          dispatch({
            type: "install-assets-error",
            message: __("Unexpected install-assets shape from the server."),
          });
          return;
        }
        dispatch({
          type: "install-assets-success",
          quickCreateURL: assets.quickCreateURL,
          externalId: assets.externalId,
          principalArn: assets.principalArn,
          requiredActions: assets.requiredActions,
        });
      },
      onError(error) {
        dispatch({
          type: "install-assets-error",
          message: formatError(
            __("Could not generate AWS install assets"),
            error as GraphQLError,
          ),
        });
      },
    });
  };

  const handleAdvanceFromStep1 = () => {
    if (!canAdvanceAwsStep(state)) return;
    dispatch({ type: "next-step" });
    handleGenerateAssets();
  };

  const handleSubmit = () => {
    if (!canAdvanceAwsStep(state) || !state.externalId) return;
    dispatch({ type: "submitting" });
    createCloudAccount({
      variables: {
        input: {
          organizationId,
          provider: "AWS",
          credentialKind: "AWS_ASSUME_ROLE",
          label: state.label,
          scopeKind: state.scopeKind,
          scopeIdentifier: state.scopeIdentifier,
          enabledAuditModules: ["ACCESS_REVIEW"],
          awsRoleArn: state.roleArn,
          // externalId is a non-secret correlation value already persisted
          // server-side and echoed back in the install-assets response.
          awsExternalId: state.externalId,
        },
        connections: [connectionId],
      },
      onCompleted(response, errors) {
        if (errors?.length) {
          dispatch({
            type: "submit-error",
            message: errors.map(e => e.message).join(", "),
          });
          return;
        }
        const probeError = response.createCloudAccount.lastProbeError;
        if (response.createCloudAccount.verifyStatus === "VERIFIED") {
          toast({
            title: __("Success"),
            description: __("AWS account connected and verified."),
            variant: "success",
          });
        } else {
          toast({
            title: __("Connected, verification pending"),
            description:
              probeError
              ?? __(
                "AWS verification did not complete; you can retry from the row actions.",
              ),
            variant: "warning",
          });
        }
        onComplete();
      },
      onError(error) {
        dispatch({
          type: "submit-error",
          message: formatError(
            __("Could not connect AWS account"),
            error as GraphQLError,
          ),
        });
      },
    });
  };

  return (
    <div className="space-y-6">
      <ol className="flex items-center gap-2 text-xs text-txt-tertiary">
        <li
          className={state.step === 1 ? "font-semibold text-txt-primary" : ""}
        >
          {__("1. Scope")}
        </li>
        <li>•</li>
        <li
          className={state.step === 2 ? "font-semibold text-txt-primary" : ""}
        >
          {__("2. Launch CloudFormation")}
        </li>
        <li>•</li>
        <li
          className={state.step === 3 ? "font-semibold text-txt-primary" : ""}
        >
          {__("3. Paste role ARN")}
        </li>
      </ol>

      {state.errorMessage && (
        <div className="rounded-md border border-border-danger bg-danger p-3 text-sm text-txt-danger">
          {state.errorMessage}
        </div>
      )}

      {state.step === 1 && (
        <div className="space-y-4">
          <Field
            label={__("Label")}
            placeholder={__("e.g. Production AWS")}
            value={state.label}
            onValueChange={(value: string) =>
              dispatch({ type: "set-label", value })}
          />
          <Field
            label={__("Region")}
            help={__("AWS region where the read-only role will be assumed.")}
            value={state.region}
            onValueChange={(value: string) =>
              dispatch({ type: "set-region", value })}
          />
          <Field
            label={__("AWS account ID")}
            placeholder="111122223333"
            value={state.scopeIdentifier}
            onValueChange={(value: string) =>
              dispatch({ type: "set-scope-identifier", value })}
          />
        </div>
      )}

      {state.step === 2 && (
        <div className="space-y-4">
          <p className="text-sm text-txt-secondary">
            {__(
              "Click the button below to open AWS CloudFormation in a new tab. The Quick-Create stack will provision a read-only role that Probo will assume.",
            )}
          </p>
          {isGenerating && (
            <p className="text-sm text-txt-tertiary">
              {__("Generating install assets…")}
            </p>
          )}
          {state.quickCreateURL && (
            <Button
              icon={IconArrowLink}
              onClick={() => safeOpenUrl(state.quickCreateURL!)}
            >
              {__("Launch CloudFormation Quick-Create")}
            </Button>
          )}
          {state.externalId && (
            <div className="rounded-md border border-border-low p-3 space-y-2">
              <h4 className="text-sm font-medium">{__("External ID")}</h4>
              <code className="block text-xs font-mono break-all">
                {state.externalId}
              </code>
              <p className="text-xs text-txt-tertiary">
                {__(
                  "This is pre-filled in the Quick-Create stack parameters. Echoed back when you submit the role ARN.",
                )}
              </p>
            </div>
          )}
          {state.requiredActions.length > 0 && (
            <div className="rounded-md border border-border-low p-3 space-y-2">
              <h4 className="text-sm font-medium flex items-center gap-1">
                <IconCircleInfo size={14} />
                {__("Read-only IAM actions used")}
              </h4>
              <ul className="text-xs font-mono space-y-1">
                {state.requiredActions.map(action => (
                  <li key={action}>{action}</li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}

      {state.step === 3 && (
        <div className="space-y-4">
          <p className="text-sm text-txt-secondary">
            {__(
              "Once the CloudFormation stack finishes (Status = CREATE_COMPLETE), copy the role ARN from the Outputs tab and paste it below.",
            )}
          </p>
          <Field
            label={__("Role ARN")}
            placeholder="arn:aws:iam::111122223333:role/probo-readonly"
            value={state.roleArn}
            onValueChange={(value: string) =>
              dispatch({ type: "set-role-arn", value })}
            autoComplete="off"
            spellCheck={false}
          />
        </div>
      )}

      <footer className="flex justify-between">
        <Button
          variant="tertiary"
          onClick={
            state.step === 1
              ? onBack
              : () => dispatch({ type: "previous-step" })
          }
          disabled={state.submitting}
        >
          {state.step === 1 ? __("Back") : __("Previous")}
        </Button>
        {state.step < 3 && (
          <Button
            disabled={!canAdvanceAwsStep(state) || isGenerating}
            onClick={
              state.step === 1
                ? handleAdvanceFromStep1
                : () => dispatch({ type: "next-step" })
            }
          >
            {__("Continue")}
          </Button>
        )}
        {state.step === 3 && (
          <Button disabled={!canAdvanceAwsStep(state)} onClick={handleSubmit}>
            {state.submitting ? __("Connecting…") : __("Connect AWS account")}
          </Button>
        )}
      </footer>
    </div>
  );
}
