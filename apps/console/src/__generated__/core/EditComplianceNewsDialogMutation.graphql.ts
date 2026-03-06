/**
 * @generated SignedSource<<61363063fd90649e7bdd20e3794c4cd6>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type MailingListUpdateStatus = "DRAFT" | "ENQUEUED" | "PROCESSING" | "SENT";
export type UpdateMailingListUpdateInput = {
  body: string;
  id: string;
  title: string;
};
export type EditComplianceNewsDialogMutation$variables = {
  input: UpdateMailingListUpdateInput;
};
export type EditComplianceNewsDialogMutation$data = {
  readonly updateMailingListUpdate: {
    readonly mailingListUpdate: {
      readonly body: string;
      readonly id: string;
      readonly status: MailingListUpdateStatus;
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
    "concreteType": "UpdateMailingListUpdatePayload",
    "kind": "LinkedField",
    "name": "updateMailingListUpdate",
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
    "cacheID": "c8bdc40008b9f647adb3428f9d56b5c3",
    "id": null,
    "metadata": {},
    "name": "EditComplianceNewsDialogMutation",
    "operationKind": "mutation",
    "text": "mutation EditComplianceNewsDialogMutation(\n  $input: UpdateMailingListUpdateInput!\n) {\n  updateMailingListUpdate(input: $input) {\n    mailingListUpdate {\n      id\n      title\n      body\n      status\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "2ecc704698ffaddc66d6498aba1500e1";

export default node;
