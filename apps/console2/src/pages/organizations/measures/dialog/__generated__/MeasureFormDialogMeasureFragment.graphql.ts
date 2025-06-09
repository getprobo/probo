/**
 * @generated SignedSource<<ad3cb8a5e721cd1d83d80540320b8313>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type MeasureState = "IMPLEMENTED" | "IN_PROGRESS" | "NOT_APPLICABLE" | "NOT_STARTED";
import { FragmentRefs } from "relay-runtime";
export type MeasureFormDialogMeasureFragment$data = {
  readonly category: string;
  readonly description: string;
  readonly id: string;
  readonly name: string;
  readonly state: MeasureState;
  readonly " $fragmentType": "MeasureFormDialogMeasureFragment";
};
export type MeasureFormDialogMeasureFragment$key = {
  readonly " $data"?: MeasureFormDialogMeasureFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"MeasureFormDialogMeasureFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "MeasureFormDialogMeasureFragment",
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
      "name": "description",
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
      "name": "category",
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

(node as any).hash = "4266d43b90952814d68b5425f65c10be";

export default node;
