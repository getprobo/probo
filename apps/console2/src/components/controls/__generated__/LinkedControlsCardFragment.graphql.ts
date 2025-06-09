/**
 * @generated SignedSource<<c922ab2420ac172a8d9c18948fc4f006>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type LinkedControlsCardFragment$data = {
  readonly framework: {
    readonly name: string;
  };
  readonly id: string;
  readonly name: string;
  readonly referenceId: string;
  readonly " $fragmentType": "LinkedControlsCardFragment";
};
export type LinkedControlsCardFragment$key = {
  readonly " $data"?: LinkedControlsCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"LinkedControlsCardFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "LinkedControlsCardFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "id",
      "storageKey": null
    },
    (v0/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "referenceId",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "Framework",
      "kind": "LinkedField",
      "name": "framework",
      "plural": false,
      "selections": [
        (v0/*: any*/)
      ],
      "storageKey": null
    }
  ],
  "type": "Control",
  "abstractKey": null
};
})();

(node as any).hash = "1b72ca232511c698e0cbdc84cbc81c02";

export default node;
