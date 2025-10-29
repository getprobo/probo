/**
 * @generated SignedSource<<552529ec3c732ce161f2dcb4a16179b9>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type SettingsPageFragment$data = {
  readonly id: string;
  readonly name: string;
  readonly " $fragmentSpreads": FragmentRefs<"DomainSettingsTabFragment" | "GeneralSettingsTabFragment" | "MembersSettingsTabInvitationsFragment" | "MembersSettingsTabMembershipsFragment" | "SAMLSettingsTabFragment">;
  readonly " $fragmentType": "SettingsPageFragment";
};
export type SettingsPageFragment$key = {
  readonly " $data"?: SettingsPageFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"SettingsPageFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "SettingsPageFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "id",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "name",
      "storageKey": null
    },
    {
      "args": null,
      "kind": "FragmentSpread",
      "name": "GeneralSettingsTabFragment"
    },
    {
      "args": null,
      "kind": "FragmentSpread",
      "name": "MembersSettingsTabMembershipsFragment"
    },
    {
      "args": null,
      "kind": "FragmentSpread",
      "name": "MembersSettingsTabInvitationsFragment"
    },
    {
      "args": null,
      "kind": "FragmentSpread",
      "name": "DomainSettingsTabFragment"
    },
    {
      "args": null,
      "kind": "FragmentSpread",
      "name": "SAMLSettingsTabFragment"
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "4f0ec089ac8ee79935eb56c22de31eca";

export default node;
