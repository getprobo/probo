/**
 * @generated SignedSource<<d1b0f29871255db02a54871ab1957a10>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentStatus = "DRAFT" | "PUBLISHED";
import { FragmentRefs } from "relay-runtime";
export type DocumentVersionsDropdownItemFragment$data = {
  readonly id: string;
  readonly publishedAt: string | null | undefined;
  readonly status: DocumentStatus;
  readonly updatedAt: string;
  readonly version: number;
  readonly " $fragmentType": "DocumentVersionsDropdownItemFragment";
};
export type DocumentVersionsDropdownItemFragment$key = {
  readonly " $data"?: DocumentVersionsDropdownItemFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentVersionsDropdownItemFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentVersionsDropdownItemFragment",
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

(node as any).hash = "923dfbcf3e171fed5ea6509a43c1b4ce";

export default node;
