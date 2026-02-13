/**
 * @generated SignedSource<<7464a4e3e5b2b5b72a6ecda47b5dace7>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentClassification = "CONFIDENTIAL" | "INTERNAL" | "PUBLIC" | "SECRET";
export type DocumentStatus = "DRAFT" | "PUBLISHED";
import { FragmentRefs } from "relay-runtime";
export type DocumentLayoutDrawer_versionFragment$data = {
  readonly classification: DocumentClassification;
  readonly id: string;
  readonly publishedAt: string | null | undefined;
  readonly status: DocumentStatus;
  readonly updatedAt: string;
  readonly version: number;
  readonly " $fragmentType": "DocumentLayoutDrawer_versionFragment";
};
export type DocumentLayoutDrawer_versionFragment$key = {
  readonly " $data"?: DocumentLayoutDrawer_versionFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentLayoutDrawer_versionFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentLayoutDrawer_versionFragment",
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
      "name": "classification",
      "storageKey": null
    },
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
      "name": "updatedAt",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "publishedAt",
      "storageKey": null
    }
  ],
  "type": "DocumentVersion",
  "abstractKey": null
};

(node as any).hash = "ebe62fad4b4044be35adb9ba4979a45a";

export default node;
