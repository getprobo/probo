/**
 * @generated SignedSource<<71619b1be320df410e3ebe2157b11ed4>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteOrganizationInput = {
  organizationId: string;
};
export type GeneralSettingsPage_deleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteOrganizationInput;
};
export type GeneralSettingsPage_deleteMutation$data = {
  readonly deleteOrganization: {
    readonly deletedOrganizationId: string;
  } | null | undefined;
};
export type GeneralSettingsPage_deleteMutation = {
  response: GeneralSettingsPage_deleteMutation$data;
  variables: GeneralSettingsPage_deleteMutation$variables;
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
  "kind": "ScalarField",
  "name": "deletedOrganizationId",
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
    "name": "GeneralSettingsPage_deleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteOrganizationPayload",
        "kind": "LinkedField",
        "name": "deleteOrganization",
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
    "name": "GeneralSettingsPage_deleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteOrganizationPayload",
        "kind": "LinkedField",
        "name": "deleteOrganization",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedOrganizationId",
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
    "cacheID": "53d34f4a9fe61e454f0ca80e8f34d88a",
    "id": null,
    "metadata": {},
    "name": "GeneralSettingsPage_deleteMutation",
    "operationKind": "mutation",
    "text": "mutation GeneralSettingsPage_deleteMutation(\n  $input: DeleteOrganizationInput!\n) {\n  deleteOrganization(input: $input) {\n    deletedOrganizationId\n  }\n}\n"
  }
};
})();

(node as any).hash = "c473475817881aaa4adee8f183faf580";

export default node;
