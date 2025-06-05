/**
 * @generated SignedSource<<ec6bd0c6fff930b82216d940e83d8ef0>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type FrameworksPageCardFragment$data = {
  readonly description: string;
  readonly id: string;
  readonly name: string;
  readonly " $fragmentType": "FrameworksPageCardFragment";
};
export type FrameworksPageCardFragment$key = {
  readonly " $data"?: FrameworksPageCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"FrameworksPageCardFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "FrameworksPageCardFragment",
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
      "name": "description",
      "storageKey": null
    }
  ],
  "type": "Framework",
  "abstractKey": null
};

(node as any).hash = "4481f380673963e72f88071031e37d14";

export default node;
