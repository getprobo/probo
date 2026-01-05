/**
 * @generated SignedSource<<17691f5400c05fd0281c9320b9fb86b9>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type SCIMConfigurationFragment$data = {
  readonly endpointUrl: string;
  readonly id: string;
  readonly " $fragmentType": "SCIMConfigurationFragment";
};
export type SCIMConfigurationFragment$key = {
  readonly " $data"?: SCIMConfigurationFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"SCIMConfigurationFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "SCIMConfigurationFragment",
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
      "name": "endpointUrl",
      "storageKey": null
    }
  ],
  "type": "SCIMConfiguration",
  "abstractKey": null
};

(node as any).hash = "5bd76da1abe24699f9895857cda748b1";

export default node;
