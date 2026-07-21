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

import { faviconUrl } from "@probo/helpers";
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
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();
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
          <Th>{t("linkedThirdPartiesCard.columns.name")}</Th>
          <Th>{t("linkedThirdPartiesCard.columns.category")}</Th>
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
              {t("linkedThirdPartiesCard.empty")}
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
              {t("linkedThirdPartiesCard.actions.link")}
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
  const { t } = useTranslation();
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
            {t("linkedThirdPartiesCard.actions.unlink")}
          </Button>
        </Td>
      )}
    </Tr>
  );
}
