/**
 * @generated SignedSource<<929e2bdd729e8b9654b27cd9e9dbccb2>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type PolicyStatus = "DRAFT" | "PUBLISHED" | "%future added value";
import { FragmentRefs } from "relay-runtime";
export type PolicyVersionHistoryDialogFragment$data = {
  readonly changelog: string;
  readonly content: string;
  readonly id: string;
  readonly publishedAt: any | null | undefined;
  readonly publishedBy: {
    readonly fullName: string;
  } | null | undefined;
  readonly status: PolicyStatus;
  readonly updatedAt: any;
  readonly version: number;
  readonly " $fragmentType": "PolicyVersionHistoryDialogFragment";
};
export type PolicyVersionHistoryDialogFragment$key = {
  readonly " $data"?: PolicyVersionHistoryDialogFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"PolicyVersionHistoryDialogFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "PolicyVersionHistoryDialogFragment",
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
  "type": "PolicyVersion",
  "abstractKey": null
};

(node as any).hash = "63ce12a876b10889d9ff2e4ee20af968";

export default node;
