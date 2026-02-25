/**
 * @generated SignedSource<<033e59bdf3d420b00f45cccf94f6155b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type OrganizationSidebar_unsubscribeFromNewsletterMutation$variables = Record<PropertyKey, never>;
export type OrganizationSidebar_unsubscribeFromNewsletterMutation$data = {
  readonly unsubscribeFromNewsletter: {
    readonly trustCenter: {
      readonly id: string;
      readonly isViewerSubscribedToNewsletter: boolean;
    };
  };
};
export type OrganizationSidebar_unsubscribeFromNewsletterMutation = {
  response: OrganizationSidebar_unsubscribeFromNewsletterMutation$data;
  variables: OrganizationSidebar_unsubscribeFromNewsletterMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "UnsubscribeFromNewsletterPayload",
    "kind": "LinkedField",
    "name": "unsubscribeFromNewsletter",
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
    "name": "OrganizationSidebar_unsubscribeFromNewsletterMutation",
    "selections": (v0/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "OrganizationSidebar_unsubscribeFromNewsletterMutation",
    "selections": (v0/*: any*/)
  },
  "params": {
    "cacheID": "4e11d3f1c4879997295a49c650cb5672",
    "id": null,
    "metadata": {},
    "name": "OrganizationSidebar_unsubscribeFromNewsletterMutation",
    "operationKind": "mutation",
    "text": "mutation OrganizationSidebar_unsubscribeFromNewsletterMutation {\n  unsubscribeFromNewsletter {\n    trustCenter {\n      id\n      isViewerSubscribedToNewsletter\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "bcb94e4e8370c92d798da68dcbd695e6";

export default node;
