/**
 * @generated SignedSource<<458f30b721ecd9df210621a8853008c3>>
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
export type NewComplianceUpdateDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateMailingListUpdateInput;
};
export type NewComplianceUpdateDialogMutation$data = {
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
export type NewComplianceUpdateDialogMutation = {
  response: NewComplianceUpdateDialogMutation$data;
  variables: NewComplianceUpdateDialogMutation$variables;
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
    "name": "NewComplianceUpdateDialogMutation",
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
    "name": "NewComplianceUpdateDialogMutation",
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
    "cacheID": "33fc392100d93803de3a189e75961a75",
    "id": null,
    "metadata": {},
    "name": "NewComplianceUpdateDialogMutation",
    "operationKind": "mutation",
    "text": "mutation NewComplianceUpdateDialogMutation(\n  $input: CreateMailingListUpdateInput!\n) {\n  createMailingListUpdate(input: $input) {\n    mailingListUpdate {\n      id\n      title\n      body\n      status\n      createdAt\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "bf02b571efaf38d4f4a8fde50b311832";

export default node;
