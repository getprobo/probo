/**
 * @generated SignedSource<<06f02e43b64c316083067f5b3978a08a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type OrganizationDropdownFragment$data = {
  readonly name: string;
  readonly " $fragmentType": "OrganizationDropdownFragment";
};
export type OrganizationDropdownFragment$key = {
  readonly " $data"?: OrganizationDropdownFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"OrganizationDropdownFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "OrganizationDropdownFragment",
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

(node as any).hash = "9f19581f86736e6345912284668a8f25";

export default node;
