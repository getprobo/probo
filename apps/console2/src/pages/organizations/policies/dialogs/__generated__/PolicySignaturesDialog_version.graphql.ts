/**
 * @generated SignedSource<<efbbec863b7456c271e408fb553f9ea1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type PolicyStatus = "DRAFT" | "PUBLISHED";
import { FragmentRefs } from "relay-runtime";
export type PolicySignaturesDialog_version$data = {
  readonly publishedAt: any | null | undefined;
  readonly status: PolicyStatus;
  readonly updatedAt: any;
  readonly version: number;
  readonly " $fragmentType": "PolicySignaturesDialog_version";
};
export type PolicySignaturesDialog_version$key = {
  readonly " $data"?: PolicySignaturesDialog_version$data;
  readonly " $fragmentSpreads": FragmentRefs<"PolicySignaturesDialog_version">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "PolicySignaturesDialog_version",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "version",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "status",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "publishedAt",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "updatedAt",
      "storageKey": null
    }
  ],
  "type": "PolicyVersion",
  "abstractKey": null
};

(node as any).hash = "cdec8e3244fd6329a86b5825a4133c44";

export default node;
