/**
 * @generated SignedSource<<9e7dea0ad7cddb603d16a2fbbe71e4dc>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type VersionRowFragment$data = {
  readonly id: string;
  readonly publishedAt: string | null | undefined;
  readonly signed: boolean;
  readonly version: number;
  readonly " $fragmentType": "VersionRowFragment";
};
export type VersionRowFragment$key = {
  readonly " $data"?: VersionRowFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"VersionRowFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "VersionRowFragment",
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
      "name": "signed",
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

(node as any).hash = "421d770febb7c38b46562e03ecd95a41";

export default node;
