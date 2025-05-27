/**
 * @generated SignedSource<<fbc31c6da75cbb26f402b2c88664950a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type PeopleGraphQuery$variables = {
  organizationId: string;
};
export type PeopleGraphQuery$data = {
  readonly organization: {
    readonly peoples?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly fullName: string;
          readonly id: string;
          readonly primaryEmailAddress: string;
        };
      }>;
    };
  };
};
export type PeopleGraphQuery = {
  response: PeopleGraphQuery$data;
  variables: PeopleGraphQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "organizationId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "organizationId"
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
  "kind": "InlineFragment",
  "selections": [
    {
      "alias": null,
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 100
        },
        {
          "kind": "Literal",
          "name": "orderBy",
          "value": {
            "direction": "ASC",
            "field": "CREATED_AT"
          }
        }
      ],
      "concreteType": "PeopleConnection",
      "kind": "LinkedField",
      "name": "peoples",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "PeopleEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "People",
              "kind": "LinkedField",
              "name": "node",
              "plural": false,
              "selections": [
                (v2/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "fullName",
                  "storageKey": null
                },
                {
                  "alias": null,
                  "args": null,
                  "kind": "ScalarField",
                  "name": "primaryEmailAddress",
                  "storageKey": null
                }
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "peoples(first:100,orderBy:{\"direction\":\"ASC\",\"field\":\"CREATED_AT\"})"
    }
  ],
  "type": "Organization",
  "abstractKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "PeopleGraphQuery",
    "selections": [
      {
        "alias": "organization",
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v3/*: any*/)
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
    "name": "PeopleGraphQuery",
    "selections": [
      {
        "alias": "organization",
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
          (v3/*: any*/),
          (v2/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "cf76244f6aaa0c573bdeb7bf118ad2a3",
    "id": null,
    "metadata": {},
    "name": "PeopleGraphQuery",
    "operationKind": "query",
    "text": "query PeopleGraphQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      peoples(first: 100, orderBy: {direction: ASC, field: CREATED_AT}) {\n        edges {\n          node {\n            id\n            fullName\n            primaryEmailAddress\n          }\n        }\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "11220f45320400c451abeaf7479a8e47";

export default node;
