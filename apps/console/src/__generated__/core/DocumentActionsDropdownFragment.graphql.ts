/**
 * @generated SignedSource<<b80ec4401c50c0af90c48812a73b76b5>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentActionsDropdownFragment$data = {
  readonly canDelete: boolean;
  readonly canUpdate: boolean;
  readonly id: string;
  readonly title: string;
  readonly versions: {
    readonly __id: string;
    readonly totalCount: number;
  };
  readonly " $fragmentSpreads": FragmentRefs<"UpdateVersionDialogFragment">;
  readonly " $fragmentType": "DocumentActionsDropdownFragment";
};
export type DocumentActionsDropdownFragment$key = {
  readonly " $data"?: DocumentActionsDropdownFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentActionsDropdownFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentActionsDropdownFragment",
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
      "name": "title",
      "storageKey": null
    },
    {
      "alias": "canUpdate",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:document:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:document:update\")"
    },
    {
      "alias": "canDelete",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:document:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:document:delete\")"
    },
    {
      "alias": null,
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 20
        }
      ],
      "concreteType": "DocumentVersionConnection",
      "kind": "LinkedField",
      "name": "versions",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "totalCount",
          "storageKey": null
        },
        {
          "kind": "ClientExtension",
          "selections": [
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "__id",
              "storageKey": null
            }
          ]
        }
      ],
      "storageKey": "versions(first:20)"
    },
    {
      "args": null,
      "kind": "FragmentSpread",
      "name": "UpdateVersionDialogFragment"
    }
  ],
  "type": "Document",
  "abstractKey": null
};

(node as any).hash = "bc3632812ef26d793478c01bd41e7dbe";

export default node;
