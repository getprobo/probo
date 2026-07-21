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

import { formatDate, formatError } from "@probo/helpers";
import {
  Badge,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Td,
  Tr,
  useToast,
} from "@probo/ui";
import * as Popover from "@radix-ui/react-popover";
import { useTranslation } from "react-i18next";
import { ConnectionHandler, graphql, useFragment, useMutation } from "react-relay";

import type { OAuthTokenRowFragment$key } from "#/__generated__/iam/OAuthTokenRowFragment.graphql";
import type { OAuthTokenRowRevokeMutation } from "#/__generated__/iam/OAuthTokenRowRevokeMutation.graphql";

import { formatApiScopeLabel } from "./scopeLabels";

const VISIBLE_SCOPE_COUNT = 3;

const fragment = graphql`
  fragment OAuthTokenRowFragment on OAuth2AccessToken {
    id
    name
    scopes
    expiresAt
    createdAt
    canDelete: permission(action: "iam:oauth2-access-token:delete")
  }
`;

const revokeMutation = graphql`
  mutation OAuthTokenRowRevokeMutation(
    $input: RevokeOAuth2AccessTokenInput!
    $connections: [ID!]!
  ) {
    revokeOAuth2AccessToken(input: $input) {
      oauth2AccessTokenId @deleteEdge(connections: $connections)
    }
  }
`;

export function OAuthTokenRow(props: {
  tokenKey: OAuthTokenRowFragment$key;
  identityId: string;
}) {
  const { tokenKey, identityId } = props;
  const { t } = useTranslation();
  const { toast } = useToast();

  const token = useFragment(fragment, tokenKey);
  const visibleScopes = token.scopes.slice(0, VISIBLE_SCOPE_COUNT);
  const hiddenScopes = token.scopes.slice(VISIBLE_SCOPE_COUNT);

  const [revoke, isRevoking] = useMutation<OAuthTokenRowRevokeMutation>(revokeMutation);

  const handleRevoke = () => {
    const connectionID = ConnectionHandler.getConnectionID(
      identityId,
      "OAuthTokensPage_oauth2AccessTokens",
    );

    revoke({
      variables: {
        input: { oauth2AccessTokenId: token.id },
        connections: [connectionID],
      },
      onCompleted: (_response, errors) => {
        if (errors?.length) {
          toast({
            title: t("common.error"),
            description: formatError(
              t("oauthTokenRow.errors.revoke"),
              errors,
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("common.success"), description: t("oauthTokenRow.messages.revoked"),
          variant: "success",
        });
      },
      onError: (error) => {
        toast({
          title: t("common.error"), description: formatError(t("oauthTokenRow.errors.revoke"), error),
          variant: "error",
        });
      },
    });
  };

  return (
    <Tr>
      <Td className="font-medium">{token.name}</Td>
      <Td noLink className="max-w-xs">
        <div className="flex flex-wrap gap-1">
          {visibleScopes.map(scope => (
            <Badge key={scope} variant="neutral" className="text-xs">
              {formatApiScopeLabel(scope)}
            </Badge>
          ))}
          {hiddenScopes.length > 0 && (
            <Popover.Root>
              <Popover.Trigger asChild>
                <button type="button" className="inline-flex">
                  <Badge variant="neutral" className="text-xs cursor-pointer">
                    +
                    {hiddenScopes.length}
                  </Badge>
                </button>
              </Popover.Trigger>
              <Popover.Portal>
                <Popover.Content
                  className="z-50 rounded-md border bg-level-0 p-3 shadow-md max-w-sm"
                  sideOffset={4}
                  align="start"
                >
                  <div className="flex flex-wrap gap-1">
                    {hiddenScopes.map(scope => (
                      <Badge key={scope} variant="neutral" className="text-xs">
                        {formatApiScopeLabel(scope)}
                      </Badge>
                    ))}
                  </div>
                </Popover.Content>
              </Popover.Portal>
            </Popover.Root>
          )}
        </div>
      </Td>
      <Td>{formatDate(token.createdAt)}</Td>
      <Td>{formatDate(token.expiresAt)}</Td>
      <Td>
        {token.canDelete && (
          <Dialog
            title={t("oauthTokenRow.revoke.title")}
            trigger={(
              <Button variant="danger" disabled={isRevoking}>
                {t("oauthTokenRow.actions.revoke")}
              </Button>
            )}
          >
            <DialogContent padded>
              <p>
                {t("oauthTokenRow.revoke.description")}
              </p>
            </DialogContent>
            <DialogFooter>
              <Button variant="danger" onClick={handleRevoke} disabled={isRevoking}>
                {t("oauthTokenRow.actions.revokeToken")}
              </Button>
            </DialogFooter>
          </Dialog>
        )}
      </Td>
    </Tr>
  );
}
