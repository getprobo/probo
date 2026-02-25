/**
 * @generated SignedSource<<61ab320d87e6fb43d245dc070871a160>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type MailingListSubscriberStatus = "CONFIRMED" | "PENDING";
export type CreateMailingListSubscriberInput = {
  email: string;
  fullName: string;
  mailingListId: string;
};
export type NewCompliancePageSubscriberDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateMailingListSubscriberInput;
};
export type NewCompliancePageSubscriberDialogMutation$data = {
  readonly createMailingListSubscriber: {
    readonly mailingListSubscriberEdge: {
      readonly cursor: string;
      readonly node: {
        readonly createdAt: string;
        readonly email: string;
        readonly fullName: string;
        readonly id: string;
        readonly status: MailingListSubscriberStatus;
      };
    };
  };
};
export type NewCompliancePageSubscriberDialogMutation = {
  response: NewCompliancePageSubscriberDialogMutation$data;
  variables: NewCompliancePageSubscriberDialogMutation$variables;
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
  "concreteType": "MailingListSubscriberEdge",
  "kind": "LinkedField",
  "name": "mailingListSubscriberEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "cursor",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "MailingListSubscriber",
      "kind": "LinkedField",
      "name": "node",
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
          "name": "fullName",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "email",
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
        }
      ],
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
    "name": "NewCompliancePageSubscriberDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateMailingListSubscriberPayload",
        "kind": "LinkedField",
        "name": "createMailingListSubscriber",
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
    "name": "NewCompliancePageSubscriberDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateMailingListSubscriberPayload",
        "kind": "LinkedField",
        "name": "createMailingListSubscriber",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "mailingListSubscriberEdge",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "6029b34056e1e0477db8bec0080b4ad1",
    "id": null,
    "metadata": {},
    "name": "NewCompliancePageSubscriberDialogMutation",
    "operationKind": "mutation",
    "text": "mutation NewCompliancePageSubscriberDialogMutation(\n  $input: CreateMailingListSubscriberInput!\n) {\n  createMailingListSubscriber(input: $input) {\n    mailingListSubscriberEdge {\n      cursor\n      node {\n        id\n        fullName\n        email\n        status\n        createdAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "fae262a9df20589ff2923f631692a6ce";

export default node;
