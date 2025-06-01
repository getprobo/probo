/**
 * @generated SignedSource<<dc19ab6bedc70a5b4546a2edb922b97f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentStatus = "DRAFT" | "PUBLISHED";
import { FragmentRefs } from "relay-runtime";
export type DocumentSignaturesDialog_version$data = {
  readonly publishedAt: any | null | undefined;
  readonly status: DocumentStatus;
  readonly updatedAt: any;
  readonly version: number;
  readonly " $fragmentType": "DocumentSignaturesDialog_version";
};
export type DocumentSignaturesDialog_version$key = {
  readonly " $data"?: DocumentSignaturesDialog_version$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentSignaturesDialog_version">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentSignaturesDialog_version",
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
  "type": "DocumentVersion",
  "abstractKey": null
};

(node as any).hash = "564746ce584a834e25a5a341155d6b2c";

export default node;
