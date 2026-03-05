/**
 * @generated SignedSource<<d57af8192babfbd781dee133589bbfd2>>
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
export type EditComplianceNewsDialogMutation$variables = {
  input: UpdateComplianceNewsInput;
};
export type EditComplianceNewsDialogMutation$data = {
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
export type EditComplianceNewsDialogMutation = {
  response: EditComplianceNewsDialogMutation$data;
  variables: EditComplianceNewsDialogMutation$variables;
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
    "name": "EditComplianceNewsDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "EditComplianceNewsDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "1f9ee25064e1f0beedc089f6300aa64e",
    "id": null,
    "metadata": {},
    "name": "EditComplianceNewsDialogMutation",
    "operationKind": "mutation",
    "text": "mutation EditComplianceNewsDialogMutation(\n  $input: UpdateComplianceNewsInput!\n) {\n  updateComplianceNews(input: $input) {\n    complianceNews {\n      id\n      title\n      body\n      status\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "ef3db353b0883df9c7b36d3149d78b48";

export default node;
