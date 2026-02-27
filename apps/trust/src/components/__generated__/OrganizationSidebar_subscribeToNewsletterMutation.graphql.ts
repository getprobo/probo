/**
 * @generated SignedSource<<b2bd634aca663dc324e2e3990850e6df>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type OrganizationSidebar_subscribeToNewsletterMutation$variables = Record<PropertyKey, never>;
export type OrganizationSidebar_subscribeToNewsletterMutation$data = {
  readonly subscribeToNewsletter: {
    readonly trustCenter: {
      readonly id: string;
      readonly isViewerSubscribedToNewsletter: boolean;
    };
  };
};
export type OrganizationSidebar_subscribeToNewsletterMutation = {
  response: OrganizationSidebar_subscribeToNewsletterMutation$data;
  variables: OrganizationSidebar_subscribeToNewsletterMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "SubscribeToNewsletterPayload",
    "kind": "LinkedField",
    "name": "subscribeToNewsletter",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenter",
        "kind": "LinkedField",
        "name": "trustCenter",
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
            "name": "isViewerSubscribedToNewsletter",
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
    "name": "OrganizationSidebar_subscribeToNewsletterMutation",
    "selections": (v0/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "OrganizationSidebar_subscribeToNewsletterMutation",
    "selections": (v0/*: any*/)
  },
  "params": {
    "cacheID": "cf3287f9ae97a7def5aa3abee6ca84d9",
    "id": null,
    "metadata": {},
    "name": "OrganizationSidebar_subscribeToNewsletterMutation",
    "operationKind": "mutation",
    "text": "mutation OrganizationSidebar_subscribeToNewsletterMutation {\n  subscribeToNewsletter {\n    trustCenter {\n      id\n      isViewerSubscribedToNewsletter\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "072a45581e8b7d92a61b6022f84841a4";

export default node;
