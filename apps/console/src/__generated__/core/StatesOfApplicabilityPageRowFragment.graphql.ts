/**
 * @generated SignedSource<<11ddaaa561cf3f1624c74940c6516f56>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type StatesOfApplicabilityPageRowFragment$data = {
  readonly applicabilityStatementsInfo: {
    readonly totalCount: number;
  };
  readonly createdAt: string;
  readonly id: string;
  readonly name: string;
  readonly " $fragmentType": "StatesOfApplicabilityPageRowFragment";
};
export type StatesOfApplicabilityPageRowFragment$key = {
  readonly " $data"?: StatesOfApplicabilityPageRowFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"StatesOfApplicabilityPageRowFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "StatesOfApplicabilityPageRowFragment",
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
      "name": "createdAt",
      "storageKey": null
    },
    {
      "alias": "applicabilityStatementsInfo",
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 0
        }
      ],
      "concreteType": "ApplicabilityStatementConnection",
      "kind": "LinkedField",
      "name": "applicabilityStatements",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "totalCount",
          "storageKey": null
        }
      ],
      "storageKey": "applicabilityStatements(first:0)"
    }
  ],
  "type": "StateOfApplicability",
  "abstractKey": null
};

(node as any).hash = "4c08bb35ecedfd7322db019dc501d33e";

export default node;
