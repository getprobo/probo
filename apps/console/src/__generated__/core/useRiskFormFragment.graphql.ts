/**
 * @generated SignedSource<<8431443063be1d35d034a615cf97bcc4>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type RiskTreatment = "ACCEPTED" | "AVOIDED" | "MITIGATED" | "TRANSFERRED";
import { FragmentRefs } from "relay-runtime";
export type useRiskFormFragment$data = {
  readonly category: string;
  readonly description: string | null | undefined;
  readonly id: string;
  readonly inherentImpact: number;
  readonly inherentLikelihood: number;
  readonly inherentRiskScore: number;
  readonly name: string;
  readonly note: string;
  readonly owner: {
    readonly id: string;
  } | null | undefined;
  readonly residualImpact: number;
  readonly residualLikelihood: number;
  readonly residualRiskScore: number;
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
      "name": "inherentRiskScore",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "residualRiskScore",
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
      "concreteType": "Profile",
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

(node as any).hash = "d9b52f879eadba5f7b1c18dcdbf85559";

export default node;
