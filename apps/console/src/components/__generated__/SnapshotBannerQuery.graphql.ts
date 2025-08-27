/**
 * @generated SignedSource<<a11c48ca33ac17dd7eaf6eeca4a0c20d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SnapshotBannerQuery$variables = {
  snapshotId: string;
};
export type SnapshotBannerQuery$data = {
  readonly node: {
    readonly createdAt?: any;
    readonly id?: string;
    readonly name?: string;
  };
};
export type SnapshotBannerQuery = {
  response: SnapshotBannerQuery$data;
  variables: SnapshotBannerQuery$variables;
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
  "name": "createdAt",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "SnapshotBannerQuery",
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
              (v4/*: any*/)
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
    "name": "SnapshotBannerQuery",
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
              (v4/*: any*/)
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
    "cacheID": "8f21371ececf96444bb6b00e52783573",
    "id": null,
    "metadata": {},
    "name": "SnapshotBannerQuery",
    "operationKind": "query",
    "text": "query SnapshotBannerQuery(\n  $snapshotId: ID!\n) {\n  node(id: $snapshotId) {\n    __typename\n    ... on Snapshot {\n      id\n      name\n      createdAt\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "1602d5b6c0ae4a3122998e4d3419af05";

export default node;
