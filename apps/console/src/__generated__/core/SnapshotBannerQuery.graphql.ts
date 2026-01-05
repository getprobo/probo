/**
 * @generated SignedSource<<2b49935877fd64d39d0cfd649f3f01a5>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SnapshotsType = "ASSETS" | "CONTINUAL_IMPROVEMENTS" | "DATA" | "NONCONFORMITIES" | "OBLIGATIONS" | "PROCESSING_ACTIVITIES" | "RISKS" | "STATES_OF_APPLICABILITY" | "VENDORS";
export type SnapshotBannerQuery$variables = {
  snapshotId: string;
};
export type SnapshotBannerQuery$data = {
  readonly node: {
    readonly createdAt?: string;
    readonly id?: string;
    readonly name?: string;
    readonly type?: SnapshotsType;
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
  "name": "type",
  "storageKey": null
},
v5 = {
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
              (v4/*: any*/),
              (v5/*: any*/)
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
              (v4/*: any*/),
              (v5/*: any*/)
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
    "cacheID": "7e026c7c030cd116b4794260934c394f",
    "id": null,
    "metadata": {},
    "name": "SnapshotBannerQuery",
    "operationKind": "query",
    "text": "query SnapshotBannerQuery(\n  $snapshotId: ID!\n) {\n  node(id: $snapshotId) {\n    __typename\n    ... on Snapshot {\n      id\n      name\n      type\n      createdAt\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "e50a4bb5133d8e918b8b2b9d2a42f1d1";

export default node;
