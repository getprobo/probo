/**
 * @generated SignedSource<<459d0a122659971be9b24c79f47691e7>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SnapshotsType = "ASSETS" | "COMPLIANCE_REGISTRIES" | "DATA" | "NON_CONFORMITY_REGISTRIES" | "RISKS" | "VENDORS";
export type CreateSnapshotInput = {
  description?: string | null | undefined;
  name: string;
  organizationId: string;
  type: SnapshotsType;
};
export type SnapshotGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateSnapshotInput;
};
export type SnapshotGraphCreateMutation$data = {
  readonly createSnapshot: {
    readonly snapshotEdge: {
      readonly node: {
        readonly createdAt: any;
        readonly description: string | null | undefined;
        readonly id: string;
        readonly name: string;
        readonly type: SnapshotsType;
      };
    };
  };
};
export type SnapshotGraphCreateMutation = {
  response: SnapshotGraphCreateMutation$data;
  variables: SnapshotGraphCreateMutation$variables;
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
    "name": "SnapshotGraphCreateMutation",
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
    "name": "SnapshotGraphCreateMutation",
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
    "cacheID": "aa2e8ace42f6154271b5d035ae18cad2",
    "id": null,
    "metadata": {},
    "name": "SnapshotGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation SnapshotGraphCreateMutation(\n  $input: CreateSnapshotInput!\n) {\n  createSnapshot(input: $input) {\n    snapshotEdge {\n      node {\n        id\n        name\n        description\n        type\n        createdAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "97d681687efde567b46938fc0541635f";

export default node;
