/**
 * @generated SignedSource<<780834b432440afd9363100bb20f2529>>
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
    readonly id: string;
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
  "name": "id",
  "storageKey": null
},
v1 = {
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
    (v0/*: any*/),
    (v1/*: any*/),
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
        (v0/*: any*/),
        (v1/*: any*/)
      ],
      "storageKey": null
    }
  ],
  "type": "Control",
  "abstractKey": null
};
})();

(node as any).hash = "741820c65c4317cac48693023c3f9bcc";

export default node;
