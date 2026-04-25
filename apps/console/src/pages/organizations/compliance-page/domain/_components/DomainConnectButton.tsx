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

import { formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, useToast } from "@probo/ui";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { DomainConnectButtonMutation } from "#/__generated__/core/DomainConnectButtonMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const initiateDomainConnectMutation = graphql`
  mutation DomainConnectButtonMutation($input: InitiateDomainConnectInput!) {
    initiateDomainConnect(input: $input) {
      redirectUrl
    }
  }
`;

export function DomainConnectButton() {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const { toast } = useToast();

  const [initiateDomainConnect, isInitiating]
    = useMutation<DomainConnectButtonMutation>(initiateDomainConnectMutation);

  function handleClick() {
    initiateDomainConnect({
      variables: {
        input: {
          organizationId,
          continueUrl: window.location.href,
        },
      },
      onCompleted(data) {
        window.location.href = data.initiateDomainConnect.redirectUrl;
      },
      onError(error) {
        toast({
          title: __("Configuration failed"),
          description: formatError(
            __("Failed to initiate Domain Connect"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  }

  return (
    <div className="space-y-3">
      <p className="text-xs text-txt-secondary">
        {__("If your DNS is managed by a provider that supports Domain Connect (Cloudflare, GoDaddy, etc.), you can automatically create the required DNS record.")}
      </p>
      <Button
        variant="secondary"
        disabled={isInitiating}
        onClick={handleClick}
      >
        {isInitiating ? __("Redirecting...") : __("Configure with Domain Connect")}
      </Button>
    </div>
  );
}
