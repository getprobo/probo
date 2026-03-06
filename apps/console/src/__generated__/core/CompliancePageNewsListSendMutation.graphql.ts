/**
 * @generated SignedSource<<58e2cf8d47422539dd39498c0684980b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type MailingListUpdateStatus = "DRAFT" | "ENQUEUED" | "PROCESSING" | "SENT";
export type SendMailingListUpdateInput = {
  id: string;
};
export type CompliancePageNewsListSendMutation$variables = {
  input: SendMailingListUpdateInput;
};
export type CompliancePageNewsListSendMutation$data = {
  readonly sendMailingListUpdate: {
    readonly mailingListUpdate: {
      readonly body: string;
      readonly id: string;
      readonly status: MailingListUpdateStatus;
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
    "concreteType": "SendMailingListUpdatePayload",
    "kind": "LinkedField",
    "name": "sendMailingListUpdate",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "MailingListUpdate",
        "kind": "LinkedField",
        "name": "mailingListUpdate",
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
    "cacheID": "d16c954883c891dec55c97ba220feb50",
    "id": null,
    "metadata": {},
    "name": "CompliancePageNewsListSendMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageNewsListSendMutation(\n  $input: SendMailingListUpdateInput!\n) {\n  sendMailingListUpdate(input: $input) {\n    mailingListUpdate {\n      id\n      title\n      body\n      status\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "8347fcd29c858a5cf19117470ad51071";

export default node;
