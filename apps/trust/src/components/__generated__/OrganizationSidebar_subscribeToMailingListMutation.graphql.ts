/**
 * @generated SignedSource<<80e6e1e3558d126472e7affb5b7b6cbd>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type OrganizationSidebar_subscribeToMailingListMutation$variables = Record<PropertyKey, never>;
export type OrganizationSidebar_subscribeToMailingListMutation$data = {
  readonly subscribeToMailingList: {
    readonly subscription: {
      readonly createdAt: any;
      readonly email: any;
      readonly id: string;
      readonly updatedAt: any;
    };
  };
};
export type OrganizationSidebar_subscribeToMailingListMutation = {
  response: OrganizationSidebar_subscribeToMailingListMutation$data;
  variables: OrganizationSidebar_subscribeToMailingListMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "SubscribeToMailingListPayload",
    "kind": "LinkedField",
    "name": "subscribeToMailingList",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "MailingListSubscriber",
        "kind": "LinkedField",
        "name": "subscription",
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
            "name": "email",
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
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "OrganizationSidebar_subscribeToMailingListMutation",
    "selections": (v0/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "OrganizationSidebar_subscribeToMailingListMutation",
    "selections": (v0/*: any*/)
  },
  "params": {
    "cacheID": "07d99ad1645ea1e8752678fd9b5e6748",
    "id": null,
    "metadata": {},
    "name": "OrganizationSidebar_subscribeToMailingListMutation",
    "operationKind": "mutation",
    "text": "mutation OrganizationSidebar_subscribeToMailingListMutation {\n  subscribeToMailingList {\n    subscription {\n      id\n      email\n      createdAt\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "80d3e85f4a55a7de7b066959f4a457c8";

export default node;
