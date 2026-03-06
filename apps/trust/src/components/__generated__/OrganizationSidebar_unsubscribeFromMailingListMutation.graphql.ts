/**
 * @generated SignedSource<<8c1211258e4480dcaa6fdf14364cb49b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type OrganizationSidebar_unsubscribeFromMailingListMutation$variables = Record<PropertyKey, never>;
export type OrganizationSidebar_unsubscribeFromMailingListMutation$data = {
  readonly unsubscribeFromMailingList: {
    readonly deletedMailingListSubscriberId: string | null | undefined;
  };
};
export type OrganizationSidebar_unsubscribeFromMailingListMutation = {
  response: OrganizationSidebar_unsubscribeFromMailingListMutation$data;
  variables: OrganizationSidebar_unsubscribeFromMailingListMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "deletedMailingListSubscriberId",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "OrganizationSidebar_unsubscribeFromMailingListMutation",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "UnsubscribeFromMailingListPayload",
        "kind": "LinkedField",
        "name": "unsubscribeFromMailingList",
        "plural": false,
        "selections": [
          (v0/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "OrganizationSidebar_unsubscribeFromMailingListMutation",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "UnsubscribeFromMailingListPayload",
        "kind": "LinkedField",
        "name": "unsubscribeFromMailingList",
        "plural": false,
        "selections": [
          (v0/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteRecord",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedMailingListSubscriberId"
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "1c6706f94348a5426c1e57d5fe99ab82",
    "id": null,
    "metadata": {},
    "name": "OrganizationSidebar_unsubscribeFromMailingListMutation",
    "operationKind": "mutation",
    "text": "mutation OrganizationSidebar_unsubscribeFromMailingListMutation {\n  unsubscribeFromMailingList {\n    deletedMailingListSubscriberId\n  }\n}\n"
  }
};
})();

(node as any).hash = "7f3949f63166fc3bd23838a44fb94993";

export default node;
