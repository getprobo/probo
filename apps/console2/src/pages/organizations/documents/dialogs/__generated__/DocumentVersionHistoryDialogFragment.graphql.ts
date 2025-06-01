/**
 * @generated SignedSource<<dd89afd13dd17b29485a3fdaf840cd42>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentStatus = "DRAFT" | "PUBLISHED";
import { FragmentRefs } from "relay-runtime";
export type DocumentVersionHistoryDialogFragment$data = {
  readonly changelog: string;
  readonly content: string;
  readonly id: string;
  readonly publishedAt: any | null | undefined;
  readonly publishedBy: {
    readonly fullName: string;
  } | null | undefined;
  readonly status: DocumentStatus;
  readonly updatedAt: any;
  readonly version: number;
  readonly " $fragmentType": "DocumentVersionHistoryDialogFragment";
};
export type DocumentVersionHistoryDialogFragment$key = {
  readonly " $data"?: DocumentVersionHistoryDialogFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentVersionHistoryDialogFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentVersionHistoryDialogFragment",
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
      "name": "content",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "changelog",
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
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "People",
      "kind": "LinkedField",
      "name": "publishedBy",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "fullName",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "DocumentVersion",
  "abstractKey": null
};

(node as any).hash = "947253287039236aa413ee33879a6e4b";

export default node;
