/**
 * @generated SignedSource<<bdbc3097fad5d998519bb991b3a664c2>>
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
  readonly sectionTitle: string;
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
      "name": "sectionTitle",
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

(node as any).hash = "20d4049be0cad5dc766f22d9a0535625";

export default node;
