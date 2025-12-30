/**
 * @generated SignedSource<<6e092be20526b76ee767836880803c34>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type LinkedRisksCardFragment$data = {
  readonly id: string;
  readonly inherentRiskScore: number;
  readonly name: string;
  readonly residualRiskScore: number;
  readonly " $fragmentType": "LinkedRisksCardFragment";
};
export type LinkedRisksCardFragment$key = {
  readonly " $data"?: LinkedRisksCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"LinkedRisksCardFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "LinkedRisksCardFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "id",
      "storageKey": null
    },
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

(node as any).hash = "32a84445ae1cc071f56139fa44d67690";

export default node;
