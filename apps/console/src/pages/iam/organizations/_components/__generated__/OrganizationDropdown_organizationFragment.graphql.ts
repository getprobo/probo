/**
 * @generated SignedSource<<1d2465517ea2dcfa99fb3dd8757fe980>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type OrganizationDropdown_organizationFragment$data = {
  readonly name: string;
  readonly " $fragmentType": "OrganizationDropdown_organizationFragment";
};
export type OrganizationDropdown_organizationFragment$key = {
  readonly " $data"?: OrganizationDropdown_organizationFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"OrganizationDropdown_organizationFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "OrganizationDropdown_organizationFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "name",
      "storageKey": null
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "8e6d0c702f821670315b5bde55e2556a";

export default node;
