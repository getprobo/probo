/**
 * @generated SignedSource<<ff0b4fa95a5c304058379d21994833ff>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MembersPage_invitationsTotalCountFragment$data = {
  totalCount: number | null | undefined;
  readonly " $fragmentType": "MembersPage_invitationsTotalCountFragment";
};
export type MembersPage_invitationsTotalCountFragment$key = {
  readonly " $data"?: MembersPage_invitationsTotalCountFragment$data;
  readonly $updatableFragmentSpreads: FragmentRefs<"MembersPage_invitationsTotalCountFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "MembersPage_invitationsTotalCountFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "totalCount",
      "storageKey": null
    }
  ],
  "type": "InvitationConnection",
  "abstractKey": null
};

(node as any).hash = "af477ef9a8b791eee11ab26caee3d304";

export default node;
