/**
 * @generated SignedSource<<d044af0467a7f2acf6c390cdcc561c40>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type StateOfApplicabilityRowFragment$data = {
  readonly canDelete: boolean;
  readonly createdAt: string;
  readonly id: string;
  readonly name: string;
  readonly statementsInfo: {
    readonly totalCount: number;
  };
  readonly " $fragmentType": "StateOfApplicabilityRowFragment";
};
export type StateOfApplicabilityRowFragment$key = {
  readonly " $data"?: StateOfApplicabilityRowFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"StateOfApplicabilityRowFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "StateOfApplicabilityRowFragment",
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
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:state-of-applicability:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:state-of-applicability:delete\")"
    },
    {
      "alias": "statementsInfo",
      "args": null,
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
      "storageKey": null
    }
  ],
  "type": "StateOfApplicability",
  "abstractKey": null
};

(node as any).hash = "4f393828b72d1b7a4417a8eb2a5415b7";

export default node;
