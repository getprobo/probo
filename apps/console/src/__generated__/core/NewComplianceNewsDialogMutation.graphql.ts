/**
 * @generated SignedSource<<5b6f5f2bce7e5fec196dfd4cdc12627e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ComplianceNewsStatus = "DRAFT" | "SENT";
export type CreateComplianceNewsInput = {
  body: string;
  title: string;
  trustCenterId: string;
};
export type NewComplianceNewsDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateComplianceNewsInput;
};
export type NewComplianceNewsDialogMutation$data = {
  readonly createComplianceNews: {
    readonly complianceNews: {
      readonly body: string;
      readonly createdAt: string;
      readonly id: string;
      readonly status: ComplianceNewsStatus;
      readonly title: string;
      readonly updatedAt: string;
    };
  };
};
export type NewComplianceNewsDialogMutation = {
  response: NewComplianceNewsDialogMutation$data;
  variables: NewComplianceNewsDialogMutation$variables;
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
  "concreteType": "ComplianceNews",
  "kind": "LinkedField",
  "name": "complianceNews",
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
      "name": "title",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "body",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "status",
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
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "updatedAt",
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
    "name": "NewComplianceNewsDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateComplianceNewsPayload",
        "kind": "LinkedField",
        "name": "createComplianceNews",
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
    "name": "NewComplianceNewsDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateComplianceNewsPayload",
        "kind": "LinkedField",
        "name": "createComplianceNews",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "appendNode",
            "key": "",
            "kind": "LinkedHandle",
            "name": "complianceNews",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              },
              {
                "kind": "Literal",
                "name": "edgeTypeName",
                "value": "ComplianceNewsEdge"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "316251945e361b61df5f5e57e7977be1",
    "id": null,
    "metadata": {},
    "name": "NewComplianceNewsDialogMutation",
    "operationKind": "mutation",
    "text": "mutation NewComplianceNewsDialogMutation(\n  $input: CreateComplianceNewsInput!\n) {\n  createComplianceNews(input: $input) {\n    complianceNews {\n      id\n      title\n      body\n      status\n      createdAt\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "45756e07f9b412a27bc115aefabcb84e";

export default node;
