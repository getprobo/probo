/**
 * @generated SignedSource<<7314c67aea13775f772b81880c524a60>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RightsRequestState = "DONE" | "IN_PROGRESS" | "TODO";
export type RightsRequestType = "ACCESS" | "DELETION" | "PORTABILITY";
export type RightsRequestGraphNodeQuery$variables = {
  rightsRequestId: string;
};
export type RightsRequestGraphNodeQuery$data = {
  readonly node: {
    readonly actionTaken?: string | null | undefined;
    readonly contact?: string | null | undefined;
    readonly createdAt?: any;
    readonly dataSubject?: string | null | undefined;
    readonly deadline?: any | null | undefined;
    readonly details?: string | null | undefined;
    readonly id?: string;
    readonly organization?: {
      readonly id: string;
      readonly name: string;
    };
    readonly requestState?: RightsRequestState;
    readonly requestType?: RightsRequestType;
    readonly updatedAt?: any;
  };
};
export type RightsRequestGraphNodeQuery = {
  response: RightsRequestGraphNodeQuery$data;
  variables: RightsRequestGraphNodeQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "rightsRequestId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "rightsRequestId"
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
  "name": "requestType",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "requestState",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "dataSubject",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "contact",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "details",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "deadline",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "actionTaken",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "concreteType": "Organization",
  "kind": "LinkedField",
  "name": "organization",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "name",
      "storageKey": null
    }
  ],
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "RightsRequestGraphNodeQuery",
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
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v12/*: any*/)
            ],
            "type": "RightsRequest",
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
    "name": "RightsRequestGraphNodeQuery",
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
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v12/*: any*/)
            ],
            "type": "RightsRequest",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "93aa5fb44159c0ae2709c34f493bc69f",
    "id": null,
    "metadata": {},
    "name": "RightsRequestGraphNodeQuery",
    "operationKind": "query",
    "text": "query RightsRequestGraphNodeQuery(\n  $rightsRequestId: ID!\n) {\n  node(id: $rightsRequestId) {\n    __typename\n    ... on RightsRequest {\n      id\n      requestType\n      requestState\n      dataSubject\n      contact\n      details\n      deadline\n      actionTaken\n      organization {\n        id\n        name\n      }\n      createdAt\n      updatedAt\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "9f7d09c59f937a8de6cfbc0354b42d1b";

export default node;
