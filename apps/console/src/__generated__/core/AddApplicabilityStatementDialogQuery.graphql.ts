/**
 * @generated SignedSource<<0a75f5819022ec5555d7c3d8b4098e1c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type AddApplicabilityStatementDialogQuery$variables = {
  organizationId: string;
  stateOfApplicabilityId: string;
};
export type AddApplicabilityStatementDialogQuery$data = {
  readonly organization: {
    readonly controls?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly framework: {
            readonly id: string;
            readonly name: string;
          };
          readonly id: string;
          readonly name: string;
          readonly sectionTitle: string;
        };
      }>;
    };
    readonly id?: string;
  };
  readonly stateOfApplicability: {
    readonly applicabilityStatements?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly applicability: boolean;
          readonly control: {
            readonly id: string;
          };
          readonly id: string;
          readonly justification: string;
        };
      }>;
    };
    readonly id?: string;
  };
};
export type AddApplicabilityStatementDialogQuery = {
  response: AddApplicabilityStatementDialogQuery$data;
  variables: AddApplicabilityStatementDialogQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "organizationId"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "stateOfApplicabilityId"
},
v2 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "stateOfApplicabilityId"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v4 = {
  "kind": "Literal",
  "name": "first",
  "value": 10000
},
v5 = {
  "alias": null,
  "args": [
    (v4/*: any*/)
  ],
  "concreteType": "ApplicabilityStatementConnection",
  "kind": "LinkedField",
  "name": "applicabilityStatements",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "ApplicabilityStatementEdge",
      "kind": "LinkedField",
      "name": "edges",
      "plural": true,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "ApplicabilityStatement",
          "kind": "LinkedField",
          "name": "node",
          "plural": false,
          "selections": [
            (v3/*: any*/),
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "applicability",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "justification",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "concreteType": "Control",
              "kind": "LinkedField",
              "name": "control",
              "plural": false,
              "selections": [
                (v3/*: any*/)
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "storageKey": "applicabilityStatements(first:10000)"
},
v6 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "organizationId"
  }
],
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": [
    (v4/*: any*/),
    {
      "kind": "Literal",
      "name": "orderBy",
      "value": {
        "direction": "ASC",
        "field": "CREATED_AT"
      }
    }
  ],
  "concreteType": "ControlConnection",
  "kind": "LinkedField",
  "name": "controls",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "ControlEdge",
      "kind": "LinkedField",
      "name": "edges",
      "plural": true,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "Control",
          "kind": "LinkedField",
          "name": "node",
          "plural": false,
          "selections": [
            (v3/*: any*/),
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "sectionTitle",
              "storageKey": null
            },
            (v7/*: any*/),
            {
              "alias": null,
              "args": null,
              "concreteType": "Framework",
              "kind": "LinkedField",
              "name": "framework",
              "plural": false,
              "selections": [
                (v3/*: any*/),
                (v7/*: any*/)
              ],
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "storageKey": "controls(first:10000,orderBy:{\"direction\":\"ASC\",\"field\":\"CREATED_AT\"})"
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
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
    "name": "AddApplicabilityStatementDialogQuery",
    "selections": [
      {
        "alias": "stateOfApplicability",
        "args": (v2/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v5/*: any*/)
            ],
            "type": "StateOfApplicability",
            "abstractKey": null
          }
        ],
        "storageKey": null
      },
      {
        "alias": "organization",
        "args": (v6/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v8/*: any*/)
            ],
            "type": "Organization",
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
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "AddApplicabilityStatementDialogQuery",
    "selections": [
      {
        "alias": "stateOfApplicability",
        "args": (v2/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v9/*: any*/),
          (v3/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v5/*: any*/)
            ],
            "type": "StateOfApplicability",
            "abstractKey": null
          }
        ],
        "storageKey": null
      },
      {
        "alias": "organization",
        "args": (v6/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v9/*: any*/),
          (v3/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v8/*: any*/)
            ],
            "type": "Organization",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "ac6bc824dd185cea6fe2928f8292eff1",
    "id": null,
    "metadata": {},
    "name": "AddApplicabilityStatementDialogQuery",
    "operationKind": "query",
    "text": "query AddApplicabilityStatementDialogQuery(\n  $stateOfApplicabilityId: ID!\n  $organizationId: ID!\n) {\n  stateOfApplicability: node(id: $stateOfApplicabilityId) {\n    __typename\n    ... on StateOfApplicability {\n      id\n      applicabilityStatements(first: 10000) {\n        edges {\n          node {\n            id\n            applicability\n            justification\n            control {\n              id\n            }\n          }\n        }\n      }\n    }\n    id\n  }\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      id\n      controls(first: 10000, orderBy: {direction: ASC, field: CREATED_AT}) {\n        edges {\n          node {\n            id\n            sectionTitle\n            name\n            framework {\n              id\n              name\n            }\n          }\n        }\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "dd10dd30b76ae39eca5953e71f7d2d02";

export default node;
