// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
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
  const { __ } = useTranslate();
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
            title: __("Error"),
            description: formatError(
              __("Failed to revoke OAuth token."),
              errors,
            ),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("OAuth token revoked."),
          variant: "success",
        });
      },
      onError: (error) => {
        toast({
          title: __("Error"),
          description: formatError(__("Failed to revoke OAuth token."), error),
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
      <Td>{new Date(token.createdAt).toLocaleDateString()}</Td>
      <Td>{new Date(token.expiresAt).toLocaleDateString()}</Td>
      <Td>
        {token.canDelete && (
          <Dialog
            title={__("Revoke OAuth token")}
            trigger={(
              <Button variant="danger" disabled={isRevoking}>
                {__("Revoke")}
              </Button>
            )}
          >
            <DialogContent padded>
              <p>
                {__(
                  "This token will stop working immediately. This action cannot be undone.",
                )}
              </p>
            </DialogContent>
            <DialogFooter>
              <Button variant="danger" onClick={handleRevoke} disabled={isRevoking}>
                {__("Revoke token")}
              </Button>
            </DialogFooter>
          </Dialog>
        )}
      </Td>
    </Tr>
  );
}
