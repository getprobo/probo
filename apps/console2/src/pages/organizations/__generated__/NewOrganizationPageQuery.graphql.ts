/**
 * @generated SignedSource<<83489e54447230c6d654a6baa9fb3e9d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type NewOrganizationPageQuery$variables = Record<PropertyKey, never>;
export type NewOrganizationPageQuery$data = {
  readonly viewer: {
    readonly id: string;
  };
};
export type NewOrganizationPageQuery = {
  response: NewOrganizationPageQuery$data;
  variables: NewOrganizationPageQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "alias": null,
    "args": null,
    "concreteType": "Viewer",
    "kind": "LinkedField",
    "name": "viewer",
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
];
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "NewOrganizationPageQuery",
    "selections": (v0/*: any*/),
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "NewOrganizationPageQuery",
    "selections": (v0/*: any*/)
  },
  "params": {
    "cacheID": "3a624ea532d0ad916c5e4bc883867f5e",
    "id": null,
    "metadata": {},
    "name": "NewOrganizationPageQuery",
    "operationKind": "query",
    "text": "query NewOrganizationPageQuery {\n  viewer {\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "c4d28895f85836c0698352ac737720ad";

export default node;
