/**
 * @generated SignedSource<<fa19c3e1b91d31188625e48b8b6db7e9>>
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
  readonly controlId: string;
  readonly id: string;
  readonly justification: string | null | undefined;
  readonly stateOfApplicability: {
    readonly id: string;
    readonly name: string;
  };
  readonly stateOfApplicabilityId: string;
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
      "kind": "ScalarField",
      "name": "stateOfApplicabilityId",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "controlId",
      "storageKey": null
    },
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
  "type": "StateOfApplicabilityControl",
  "abstractKey": null
};
})();

(node as any).hash = "b9a9922b27f277a769025f11b0829bdc";

export default node;
