/**
 * @generated SignedSource<<365bac44d5935d4e373f6f126266584c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type RiskOverviewTabFragment$data = {
  readonly inherentImpact: number;
  readonly inherentLikelihood: number;
  readonly residualImpact: number;
  readonly residualLikelihood: number;
  readonly " $fragmentType": "RiskOverviewTabFragment";
};
export type RiskOverviewTabFragment$key = {
  readonly " $data"?: RiskOverviewTabFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"RiskOverviewTabFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "RiskOverviewTabFragment",
  "selections": [
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
    }
  ],
  "type": "Risk",
  "abstractKey": null
};

(node as any).hash = "1268614eb62d89df8b8b7153d9c72082";

export default node;
