/**
 * @generated SignedSource<<3ffcc7618d81b47e39ca5731fe75e2ed>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type usePolicyFormFragment$data = {
  readonly description: string;
  readonly id: string;
  readonly owner: {
    readonly id: string;
  };
  readonly title: string;
  readonly " $fragmentType": "usePolicyFormFragment";
};
export type usePolicyFormFragment$key = {
  readonly " $data"?: usePolicyFormFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"usePolicyFormFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "usePolicyFormFragment",
  "selections": [
    (v0/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "title",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "description",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "People",
      "kind": "LinkedField",
      "name": "owner",
      "plural": false,
      "selections": [
        (v0/*: any*/)
      ],
      "storageKey": null
    }
  ],
  "type": "Policy",
  "abstractKey": null
};
})();

(node as any).hash = "619d6845e45fb8c052a8989ea546414d";

export default node;
