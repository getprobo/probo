/**
 * @generated SignedSource<<ab309c7bbc8e33d319f6edafe017e0e9>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type LinkedStatesOfApplicabilityCardFragment$data = {
  readonly applicability: boolean;
  readonly control: {
    readonly id: string;
  };
  readonly id: string;
  readonly justification: string;
  readonly stateOfApplicability: {
    readonly id: string;
    readonly name: string;
  };
  readonly " $fragmentType": "LinkedStatesOfApplicabilityCardFragment";
};
export type LinkedStatesOfApplicabilityCardFragment$key = {
  readonly " $data"?: LinkedStatesOfApplicabilityCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"LinkedStatesOfApplicabilityCardFragment">;
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
  "name": "LinkedStatesOfApplicabilityCardFragment",
  "selections": [
    (v0/*: any*/),
    {
      "alias": null,
      "args": null,
      "concreteType": "StateOfApplicability",
      "kind": "LinkedField",
      "name": "stateOfApplicability",
      "plural": false,
      "selections": [
        (v0/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "name",
          "storageKey": null
        }
      ],
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "Control",
      "kind": "LinkedField",
      "name": "control",
      "plural": false,
      "selections": [
        (v0/*: any*/)
      ],
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "applicability",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "justification",
      "storageKey": null
    }
  ],
  "type": "ApplicabilityStatement",
  "abstractKey": null
};
})();

(node as any).hash = "92dc33d1e220b3433508e55cb05ea39a";

export default node;
