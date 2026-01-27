/**
 * @generated SignedSource<<190e32d0bc7808055f7390730de5efd4>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentActionsDropdown_documentFragment$data = {
  readonly canDelete: boolean;
  readonly canUpdate: boolean;
  readonly id: string;
  readonly title: string;
  readonly versions: {
    readonly __id: string;
    readonly totalCount: number;
  };
  readonly " $fragmentSpreads": FragmentRefs<"UpdateVersionDialogFragment">;
  readonly " $fragmentType": "DocumentActionsDropdown_documentFragment";
};
export type DocumentActionsDropdown_documentFragment$key = {
  readonly " $data"?: DocumentActionsDropdown_documentFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DocumentActionsDropdown_documentFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DocumentActionsDropdown_documentFragment",
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

(node as any).hash = "0e6d41c74914e28aea353ddef6bf7385";

export default node;
