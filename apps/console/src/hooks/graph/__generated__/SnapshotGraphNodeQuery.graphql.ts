/**
 * @generated SignedSource<<5bcc9f9b743d01fc99fd749b970dccd3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SnapshotsType = "ASSETS" | "COMPLIANCE_REGISTRIES" | "DATA" | "NON_CONFORMITY_REGISTRIES" | "RISKS" | "VENDORS";
export type SnapshotGraphNodeQuery$variables = {
  snapshotId: string;
};
export type SnapshotGraphNodeQuery$data = {
  readonly node: {
    readonly createdAt?: any;
    readonly description?: string | null | undefined;
    readonly id?: string;
    readonly name?: string;
    readonly organization?: {
      readonly id: string;
      readonly name: string;
    };
    readonly type?: SnapshotsType;
  };
};
export type SnapshotGraphNodeQuery = {
  response: SnapshotGraphNodeQuery$data;
  variables: SnapshotGraphNodeQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "snapshotId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "snapshotId"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "description",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "type",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "concreteType": "Organization",
  "kind": "LinkedField",
  "name": "organization",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    (v3/*: any*/)
  ],
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "SnapshotGraphNodeQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              (v2/*: any*/),
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/)
            ],
            "type": "Snapshot",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SnapshotGraphNodeQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "__typename",
            "storageKey": null
          },
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/)
            ],
            "type": "Snapshot",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "8dafbbd8ecc8442fa065d19bba932ce2",
    "id": null,
    "metadata": {},
    "name": "SnapshotGraphNodeQuery",
    "operationKind": "query",
    "text": "query SnapshotGraphNodeQuery(\n  $snapshotId: ID!\n) {\n  node(id: $snapshotId) {\n    __typename\n    ... on Snapshot {\n      id\n      name\n      description\n      type\n      organization {\n        id\n        name\n      }\n      createdAt\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "b687d3da56dbfa333e66f4f2c9c0c824";

export default node;
