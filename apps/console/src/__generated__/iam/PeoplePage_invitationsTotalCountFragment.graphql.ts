/**
 * @generated SignedSource<<8e60217d0397fc8690fc79915bdfdc6c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type PeoplePage_invitationsTotalCountFragment$data = {
  totalCount: number | null | undefined;
  readonly " $fragmentType": "PeoplePage_invitationsTotalCountFragment";
};
export type PeoplePage_invitationsTotalCountFragment$key = {
  readonly " $data"?: PeoplePage_invitationsTotalCountFragment$data;
  readonly $updatableFragmentSpreads: FragmentRefs<"PeoplePage_invitationsTotalCountFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "PeoplePage_invitationsTotalCountFragment",
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

(node as any).hash = "8881fd3459c738d64d70370fd317fe5d";

export default node;
