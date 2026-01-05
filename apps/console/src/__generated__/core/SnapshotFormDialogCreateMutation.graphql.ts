/**
 * @generated SignedSource<<8fe6c01dea59b2ae2d04050e3cb959d1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SnapshotsType = "ASSETS" | "CONTINUAL_IMPROVEMENTS" | "DATA" | "NONCONFORMITIES" | "OBLIGATIONS" | "PROCESSING_ACTIVITIES" | "RISKS" | "STATES_OF_APPLICABILITY" | "VENDORS";
export type CreateSnapshotInput = {
  description?: string | null | undefined;
  name: string;
  organizationId: string;
  type: SnapshotsType;
};
export type SnapshotFormDialogCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateSnapshotInput;
};
export type SnapshotFormDialogCreateMutation$data = {
  readonly createSnapshot: {
    readonly snapshotEdge: {
      readonly node: {
        readonly canDelete: boolean;
        readonly createdAt: string;
        readonly description: string | null | undefined;
        readonly id: string;
        readonly name: string;
        readonly type: SnapshotsType;
      };
    };
  };
};
export type SnapshotFormDialogCreateMutation = {
  response: SnapshotFormDialogCreateMutation$data;
  variables: SnapshotFormDialogCreateMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "concreteType": "SnapshotEdge",
  "kind": "LinkedField",
  "name": "snapshotEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "Snapshot",
      "kind": "LinkedField",
      "name": "node",
      "plural": false,
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
          "name": "description",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "type",
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
              "value": "core:snapshot:delete"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:snapshot:delete\")"
        }
      ],
      "storageKey": null
    }
  ],
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "SnapshotFormDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateSnapshotPayload",
        "kind": "LinkedField",
        "name": "createSnapshot",
        "plural": false,
        "selections": [
          (v3/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "SnapshotFormDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateSnapshotPayload",
        "kind": "LinkedField",
        "name": "createSnapshot",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "snapshotEdge",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "5165891e3f49f2b9e3d1a0b29a0f8bf6",
    "id": null,
    "metadata": {},
    "name": "SnapshotFormDialogCreateMutation",
    "operationKind": "mutation",
    "text": "mutation SnapshotFormDialogCreateMutation(\n  $input: CreateSnapshotInput!\n) {\n  createSnapshot(input: $input) {\n    snapshotEdge {\n      node {\n        id\n        name\n        description\n        type\n        createdAt\n        canDelete: permission(action: \"core:snapshot:delete\")\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "1a844eb9af4ea94b194deaf65ba068a0";

export default node;
