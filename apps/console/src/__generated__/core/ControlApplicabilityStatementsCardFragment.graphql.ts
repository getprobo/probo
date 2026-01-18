/**
 * @generated SignedSource<<20a01d29fd42966dee5ab7756cbfac74>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ControlApplicabilityStatementsCardFragment$data = {
  readonly applicability: boolean;
  readonly controlId: string;
  readonly id: string;
  readonly justification: string | null | undefined;
  readonly stateOfApplicability: {
    readonly id: string;
    readonly name: string;
  };
  readonly stateOfApplicabilityId: string;
  readonly " $fragmentType": "ControlApplicabilityStatementsCardFragment";
};
export type ControlApplicabilityStatementsCardFragment$key = {
  readonly " $data"?: ControlApplicabilityStatementsCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"ControlApplicabilityStatementsCardFragment">;
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
  "name": "ControlApplicabilityStatementsCardFragment",
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

(node as any).hash = "7ba72656cb5ec463a5b853b851d8ad7b";

export default node;
