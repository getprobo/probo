/**
 * @generated SignedSource<<622eccc63036573a8e38ffc8dd36e7d7>>
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
export type EditComplianceUpdateDialogMutation$variables = {
  input: UpdateMailingListUpdateInput;
};
export type EditComplianceUpdateDialogMutation$data = {
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
export type EditComplianceUpdateDialogMutation = {
  response: EditComplianceUpdateDialogMutation$data;
  variables: EditComplianceUpdateDialogMutation$variables;
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
    "name": "EditComplianceUpdateDialogMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "EditComplianceUpdateDialogMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "49b4c9f194a0bf6ebe01408ae6387f2e",
    "id": null,
    "metadata": {},
    "name": "EditComplianceUpdateDialogMutation",
    "operationKind": "mutation",
    "text": "mutation EditComplianceUpdateDialogMutation(\n  $input: UpdateMailingListUpdateInput!\n) {\n  updateMailingListUpdate(input: $input) {\n    mailingListUpdate {\n      id\n      title\n      body\n      status\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "8b2246ab3bddba658cc488610afcd80f";

export default node;
