/**
 * @generated SignedSource<<6cf0470d3c2e5d8d9d6ca52bf3375026>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type RiskTreatment = "ACCEPTED" | "AVOIDED" | "MITIGATED" | "TRANSFERRED" | "%future added value";
import { FragmentRefs } from "relay-runtime";
export type useRiskFormFragment$data = {
  readonly category: string;
  readonly description: string;
  readonly id: string;
  readonly inherentImpact: number;
  readonly inherentLikelihood: number;
  readonly name: string;
  readonly note: string;
  readonly owner: {
    readonly id: string;
  } | null | undefined;
  readonly residualImpact: number;
  readonly residualLikelihood: number;
  readonly treatment: RiskTreatment;
  readonly " $fragmentType": "useRiskFormFragment";
};
export type useRiskFormFragment$key = {
  readonly " $data"?: useRiskFormFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"useRiskFormFragment">;
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
  "name": "useRiskFormFragment",
  "selections": [
    (v0/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "name",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "category",
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
      "kind": "ScalarField",
      "name": "treatment",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "inherentLikelihood",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "inherentImpact",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "residualLikelihood",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "residualImpact",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "note",
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
  "type": "Risk",
  "abstractKey": null
};
})();

(node as any).hash = "7ad9386b4049af3f139d7269bb5229c2";

export default node;
