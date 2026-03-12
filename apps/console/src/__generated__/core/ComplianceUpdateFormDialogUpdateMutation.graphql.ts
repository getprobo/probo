/**
 * @generated SignedSource<<160a29eaaba4b3db2c1b7aca34164dd7>>
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
export type ComplianceUpdateFormDialogUpdateMutation$variables = {
  input: UpdateMailingListUpdateInput;
};
export type ComplianceUpdateFormDialogUpdateMutation$data = {
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
export type ComplianceUpdateFormDialogUpdateMutation = {
  response: ComplianceUpdateFormDialogUpdateMutation$data;
  variables: ComplianceUpdateFormDialogUpdateMutation$variables;
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
    "name": "ComplianceUpdateFormDialogUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ComplianceUpdateFormDialogUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "49b5feb2a9a2e157568ad2431e032425",
    "id": null,
    "metadata": {},
    "name": "ComplianceUpdateFormDialogUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation ComplianceUpdateFormDialogUpdateMutation(\n  $input: UpdateMailingListUpdateInput!\n) {\n  updateMailingListUpdate(input: $input) {\n    mailingListUpdate {\n      id\n      title\n      body\n      status\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "3737c127b41a55df71e09d64aaa3da9e";

export default node;
