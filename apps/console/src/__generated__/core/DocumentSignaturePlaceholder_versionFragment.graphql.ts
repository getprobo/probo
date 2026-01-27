/**
 * @generated SignedSource<<34ec1be8ec1087ce5163f0642009d71a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentStatus = "DRAFT" | "PUBLISHED";
import { FragmentRefs } from "relay-runtime";
export type DocumentSignaturePlaceholder_versionFragment$data = {
  readonly id: string;
  readonly status: DocumentStatus;
  readonly " $fragmentType": "DocumentSignaturePlaceholder_versionFragment";
};
export type DocumentSignaturePlaceholder_versionFragment$key = {
  readonly " $data"?: DocumentSignaturePlaceholder_versionFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentSignaturePlaceholder_versionFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentSignaturePlaceholder_versionFragment",
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
      "name": "status",
      "storageKey": null
    }
  ],
  "type": "DocumentVersion",
  "abstractKey": null
};

(node as any).hash = "935a86ea9eb392eedff1c0d0558f4aac";

export default node;
