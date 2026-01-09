/**
 * @generated SignedSource<<c14e7d1b8a5a8d23dbe52637610f13f7>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type MeasureDetailPageTasksCountQuery$variables = {
  measureId: string;
};
export type MeasureDetailPageTasksCountQuery$data = {
  readonly node: {
    readonly tasks?: {
      readonly totalCount: number;
    };
  };
};
export type MeasureDetailPageTasksCountQuery = {
  response: MeasureDetailPageTasksCountQuery$data;
  variables: MeasureDetailPageTasksCountQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "measureId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "measureId"
  }
],
v2 = {
  "kind": "InlineFragment",
  "selections": [
    {
      "alias": null,
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 0
        }
      ],
      "concreteType": "TaskConnection",
      "kind": "LinkedField",
      "name": "tasks",
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
      "storageKey": "tasks(first:0)"
    }
  ],
  "type": "Measure",
  "abstractKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "MeasureDetailPageTasksCountQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v2/*: any*/)
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
    "name": "MeasureDetailPageTasksCountQuery",
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
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "36a1d5cf32d07ed5d76e5e64e7f19005",
    "id": null,
    "metadata": {},
    "name": "MeasureDetailPageTasksCountQuery",
    "operationKind": "query",
    "text": "query MeasureDetailPageTasksCountQuery(\n  $measureId: ID!\n) {\n  node(id: $measureId) {\n    __typename\n    ... on Measure {\n      tasks(first: 0) {\n        totalCount\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "f0a8dde53529b525f3fe9aa35dee9e8a";

export default node;
