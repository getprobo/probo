/**
 * @generated SignedSource<<2cec7e8ca8209de4cb2d007ad8bfe9fa>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type MailingListUpdateStatus = "DRAFT" | "ENQUEUED" | "PROCESSING" | "SENT";
export type CreateMailingListUpdateInput = {
  body: string;
  mailingListId: string;
  title: string;
};
export type ComplianceUpdateFormDialogCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateMailingListUpdateInput;
};
export type ComplianceUpdateFormDialogCreateMutation$data = {
  readonly createMailingListUpdate: {
    readonly mailingListUpdate: {
      readonly body: string;
      readonly createdAt: string;
      readonly id: string;
      readonly status: MailingListUpdateStatus;
      readonly title: string;
      readonly updatedAt: string;
    };
  };
};
export type ComplianceUpdateFormDialogCreateMutation = {
  response: ComplianceUpdateFormDialogCreateMutation$data;
  variables: ComplianceUpdateFormDialogCreateMutation$variables;
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
    "name": "ComplianceUpdateFormDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateMailingListUpdatePayload",
        "kind": "LinkedField",
        "name": "createMailingListUpdate",
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
    "name": "ComplianceUpdateFormDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateMailingListUpdatePayload",
        "kind": "LinkedField",
        "name": "createMailingListUpdate",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependNode",
            "key": "",
            "kind": "LinkedHandle",
            "name": "mailingListUpdate",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              },
              {
                "kind": "Literal",
                "name": "edgeTypeName",
                "value": "MailingListUpdateEdge"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "52d07d51ab21154cf26f8a932aba5e77",
    "id": null,
    "metadata": {},
    "name": "ComplianceUpdateFormDialogCreateMutation",
    "operationKind": "mutation",
    "text": "mutation ComplianceUpdateFormDialogCreateMutation(\n  $input: CreateMailingListUpdateInput!\n) {\n  createMailingListUpdate(input: $input) {\n    mailingListUpdate {\n      id\n      title\n      body\n      status\n      createdAt\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "6e8c7b014c850c90bcfb876d6e504f4d";

export default node;
