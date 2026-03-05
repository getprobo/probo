/**
 * @generated SignedSource<<71f99a6bd7535e0b84e60e9a00a7a759>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ComplianceNewsStatus = "DRAFT" | "SENT";
export type UpdateComplianceNewsInput = {
  body: string;
  id: string;
  status: ComplianceNewsStatus;
  title: string;
};
export type CompliancePageNewsListSendMutation$variables = {
  input: UpdateComplianceNewsInput;
};
export type CompliancePageNewsListSendMutation$data = {
  readonly updateComplianceNews: {
    readonly complianceNews: {
      readonly body: string;
      readonly id: string;
      readonly status: ComplianceNewsStatus;
      readonly title: string;
      readonly updatedAt: string;
    };
  };
};
export type CompliancePageNewsListSendMutation = {
  response: CompliancePageNewsListSendMutation$data;
  variables: CompliancePageNewsListSendMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "UpdateComplianceNewsPayload",
    "kind": "LinkedField",
    "name": "updateComplianceNews",
    "plural": false,
    "selections": [
      {
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
            "name": "updatedAt",
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "CompliancePageNewsListSendMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CompliancePageNewsListSendMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "5aff1f911ee9f493896c8f6af429c91a",
    "id": null,
    "metadata": {},
    "name": "CompliancePageNewsListSendMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageNewsListSendMutation(\n  $input: UpdateComplianceNewsInput!\n) {\n  updateComplianceNews(input: $input) {\n    complianceNews {\n      id\n      title\n      body\n      status\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "27e89ab1fd7e60850400c81ebdb95edd";

export default node;
