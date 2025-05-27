/**
 * @generated SignedSource<<45feb21e3f613b3818af04725831a427>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type PolicyStatus = "DRAFT" | "PUBLISHED";
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
