// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { faviconUrl } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  IconPlusLarge,
  IconTrashCan,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  TrButton,
} from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { LinkedThirdPartiesCardFragment$key } from "#/__generated__/core/LinkedThirdPartiesCardFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { LinkedThirdPartiesDialog } from "./LinkedThirdPartiesDialog";

const linkedThirdPartyFragment = graphql`
  fragment LinkedThirdPartiesCardFragment on ThirdParty {
    id
    name
    category
    websiteUrl
  }
`;

type Mutation<Params> = (p: {
  variables: {
    input: {
      thirdPartyId: string;
    } & Params;
    connections: string[];
  };
}) => void;

type Props<Params> = {
  thirdParties: (LinkedThirdPartiesCardFragment$key & { id: string })[];
  params: Params;
  disabled?: boolean;
  connectionId: string;
  onAttach: Mutation<Params>;
  onDetach: Mutation<Params>;
  readOnly?: boolean;
};

export function LinkedThirdPartiesCard<Params>(props: Props<Params>) {
  const { __ } = useTranslate();
  const thirdParties = props.thirdParties;

  const onAttach = (thirdPartyId: string) => {
    props.onAttach({
      variables: {
        input: {
          thirdPartyId,
          ...props.params,
        },
        connections: [props.connectionId],
      },
    });
  };

  const onDetach = (thirdPartyId: string) => {
    props.onDetach({
      variables: {
        input: {
          thirdPartyId,
          ...props.params,
        },
        connections: [props.connectionId],
      },
    });
  };

  return (
    <Table>
      <Thead>
        <Tr>
          <Th>{__("Name")}</Th>
          <Th>{__("Category")}</Th>
          {!props.readOnly && <Th></Th>}
        </Tr>
      </Thead>
      <Tbody>
        {thirdParties.length === 0 && (
          <Tr>
            <Td
              colSpan={props.readOnly ? 2 : 3}
              className="text-center text-txt-secondary"
            >
              {__("No third parties linked")}
            </Td>
          </Tr>
        )}
        {thirdParties.map(thirdParty => (
          <ThirdPartyRow
            key={thirdParty.id}
            thirdParty={thirdParty}
            onClick={onDetach}
            readOnly={props.readOnly}
          />
        ))}
        {!props.readOnly && (
          <LinkedThirdPartiesDialog
            connectionId={props.connectionId}
            disabled={props.disabled}
            linkedThirdParties={thirdParties}
            onLink={onAttach}
            onUnlink={onDetach}
          >
            <TrButton colspan={3} icon={IconPlusLarge}>
              {__("Link third party")}
            </TrButton>
          </LinkedThirdPartiesDialog>
        )}
      </Tbody>
    </Table>
  );
}

function ThirdPartyRow(props: {
  thirdParty: LinkedThirdPartiesCardFragment$key & { id: string };
  onClick: (thirdPartyId: string) => void;
  readOnly?: boolean;
}) {
  const thirdParty = useFragment(linkedThirdPartyFragment, props.thirdParty);
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const logo = faviconUrl(thirdParty.websiteUrl);

  return (
    <Tr
      to={`/organizations/${organizationId}/third-parties/${thirdParty.id}/overview`}
    >
      <Td>
        <span className="inline-flex gap-2 items-center">
          {logo && (
            <img
              src={logo}
              alt={thirdParty.name}
              className="rounded h-5 w-5"
            />
          )}
          {thirdParty.name}
        </span>
      </Td>
      <Td>
        <Badge size="md">{thirdParty.category}</Badge>
      </Td>
      {!props.readOnly && (
        <Td noLink width={50} className="text-end">
          <Button
            variant="secondary"
            onClick={() => props.onClick(thirdParty.id)}
            icon={IconTrashCan}
          >
            {__("Unlink")}
          </Button>
        </Td>
      )}
    </Tr>
  );
}
