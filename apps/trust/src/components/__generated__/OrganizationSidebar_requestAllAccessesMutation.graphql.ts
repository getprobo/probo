/**
 * @generated SignedSource<<e1d14b17505ed1107ac96553e3d8f922>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type OrganizationSidebar_requestAllAccessesMutation$variables = Record<PropertyKey, never>;
export type OrganizationSidebar_requestAllAccessesMutation$data = {
  readonly requestAllAccesses: {
    readonly trustCenterAccess: {
      readonly id: string;
    };
  };
};
export type OrganizationSidebar_requestAllAccessesMutation = {
  response: OrganizationSidebar_requestAllAccessesMutation$data;
  variables: OrganizationSidebar_requestAllAccessesMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "RequestAccessesPayload",
    "kind": "LinkedField",
    "name": "requestAllAccesses",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenterAccess",
        "kind": "LinkedField",
        "name": "trustCenterAccess",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
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
    "name": "OrganizationSidebar_requestAllAccessesMutation",
    "selections": (v0/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "OrganizationSidebar_requestAllAccessesMutation",
    "selections": (v0/*: any*/)
  },
  "params": {
    "cacheID": "a298b2a1c2f62300ae5056743d2c2ec3",
    "id": null,
    "metadata": {},
    "name": "OrganizationSidebar_requestAllAccessesMutation",
    "operationKind": "mutation",
    "text": "mutation OrganizationSidebar_requestAllAccessesMutation {\n  requestAllAccesses {\n    trustCenterAccess {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "a22225757510c4dd097e99dcd3c066a6";

export default node;
