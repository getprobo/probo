/**
 * @generated SignedSource<<7eced414e0638e806b52e86afcb6cca4>>
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
export type SendUpdateDialogMutation$variables = {
  input: SendMailingListUpdateInput;
};
export type SendUpdateDialogMutation$data = {
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
export type SendUpdateDialogMutation = {
  response: SendUpdateDialogMutation$data;
  variables: SendUpdateDialogMutation$variables;
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
    "name": "SendUpdateDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SendUpdateDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "714c7cc9cdb06cb384690e17cdaa9958",
    "id": null,
    "metadata": {},
    "name": "SendUpdateDialogMutation",
    "operationKind": "mutation",
    "text": "mutation SendUpdateDialogMutation(\n  $input: SendMailingListUpdateInput!\n) {\n  sendMailingListUpdate(input: $input) {\n    mailingListUpdate {\n      id\n      title\n      body\n      status\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "9ec6f363395b934616b97fe50b83d328";

export default node;
