/**
 * @generated SignedSource<<2c18e7bff6b76551479def8a10cce183>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type domainSettingsPage_permissionsQuery$variables = {
  organizationId: string;
};
export type domainSettingsPage_permissionsQuery$data = {
  readonly viewer: {
    readonly canCreateCustomDomain: boolean;
    readonly canDeleteCustomDomain: boolean;
  };
};
export type domainSettingsPage_permissionsQuery = {
  response: domainSettingsPage_permissionsQuery$data;
  variables: domainSettingsPage_permissionsQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "organizationId"
  }
],
v1 = {
  "kind": "Variable",
  "name": "id",
  "variableName": "organizationId"
},
v2 = {
  "alias": "canCreateCustomDomain",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:custom-domain:create"
    },
    (v1/*: any*/)
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": null
},
v3 = {
  "alias": "canDeleteCustomDomain",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:custom-domain:delete"
    },
    (v1/*: any*/)
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "domainSettingsPage_permissionsQuery",
    "selections": [
      {
        "kind": "RequiredField",
        "field": {
          "alias": null,
          "args": null,
          "concreteType": "Identity",
          "kind": "LinkedField",
          "name": "viewer",
          "plural": false,
          "selections": [
            (v2/*: any*/),
            (v3/*: any*/)
          ],
          "storageKey": null
        },
        "action": "THROW"
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "domainSettingsPage_permissionsQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Identity",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          (v3/*: any*/),
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
    ]
  },
  "params": {
    "cacheID": "bb016bc1f53b609efc4ac93a93653d08",
    "id": null,
    "metadata": {},
    "name": "domainSettingsPage_permissionsQuery",
    "operationKind": "query",
    "text": "query domainSettingsPage_permissionsQuery(\n  $organizationId: ID!\n) {\n  viewer {\n    canCreateCustomDomain: permission(action: \"core:custom-domain:create\", id: $organizationId)\n    canDeleteCustomDomain: permission(action: \"core:custom-domain:delete\", id: $organizationId)\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "f09bdc9336bf1ad1339bb3b60987f53f";

export default node;
