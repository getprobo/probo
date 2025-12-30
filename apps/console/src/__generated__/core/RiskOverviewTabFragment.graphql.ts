/**
 * @generated SignedSource<<2b276a39b8020e7be036acb6fe4ca077>>
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
  readonly inherentRiskScore: number;
  readonly residualImpact: number;
  readonly residualLikelihood: number;
  readonly residualRiskScore: number;
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
    }
  ],
  "type": "Risk",
  "abstractKey": null
};

(node as any).hash = "ba472e37a44c06c48847501204080fb6";

export default node;
