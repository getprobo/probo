/**
 * @generated SignedSource<<50304bfb34b0581895a71491f521c273>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ControlFragment_Control$data = {
  readonly description: string;
  readonly id: string;
  readonly name: string;
  readonly sectionTitle: string;
  readonly " $fragmentType": "ControlFragment_Control";
};
export type ControlFragment_Control$key = {
  readonly " $data"?: ControlFragment_Control$data;
  readonly " $fragmentSpreads": FragmentRefs<"ControlFragment_Control">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "ControlFragment_Control",
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
      "name": "sectionTitle",
      "storageKey": null
    }
  ],
  "type": "Control",
  "abstractKey": null
};

(node as any).hash = "b04ce9de2a583d415011d88563120b09";

export default node;
