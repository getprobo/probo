/**
 * @generated SignedSource<<b7482b43ce0f4dfab470c35e806acf50>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type MeasureState = "IMPLEMENTED" | "IN_PROGRESS" | "NOT_APPLICABLE" | "NOT_STARTED";
import { FragmentRefs } from "relay-runtime";
export type LinkedMeasuresCardFragment$data = {
  readonly id: string;
  readonly name: string;
  readonly state: MeasureState;
  readonly " $fragmentType": "LinkedMeasuresCardFragment";
};
export type LinkedMeasuresCardFragment$key = {
  readonly " $data"?: LinkedMeasuresCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"LinkedMeasuresCardFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "LinkedMeasuresCardFragment",
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
      "name": "state",
      "storageKey": null
    }
  ],
  "type": "Measure",
  "abstractKey": null
};

(node as any).hash = "741a216c02732c1ff97b265ac8dbf39b";

export default node;
