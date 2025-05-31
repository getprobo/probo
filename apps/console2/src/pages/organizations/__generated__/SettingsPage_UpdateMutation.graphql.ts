/**
 * @generated SignedSource<<6dc30c35c9ce7367967d235b1a259179>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateOrganizationInput = {
  logo?: any | null | undefined;
  name?: string | null | undefined;
  organizationId: string;
};
export type SettingsPage_UpdateMutation$variables = {
  input: UpdateOrganizationInput;
};
export type SettingsPage_UpdateMutation$data = {
  readonly updateOrganization: {
    readonly organization: {
      readonly id: string;
      readonly logoUrl: string | null | undefined;
      readonly name: string;
    };
  };
};
export type SettingsPage_UpdateMutation = {
  response: SettingsPage_UpdateMutation$data;
  variables: SettingsPage_UpdateMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "UpdateOrganizationPayload",
    "kind": "LinkedField",
    "name": "updateOrganization",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Organization",
        "kind": "LinkedField",
        "name": "organization",
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
            "name": "name",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "logoUrl",
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
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "SettingsPage_UpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SettingsPage_UpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "ac8d86332fabe20dd7ea8fef4207b75e",
    "id": null,
    "metadata": {},
    "name": "SettingsPage_UpdateMutation",
    "operationKind": "mutation",
    "text": "mutation SettingsPage_UpdateMutation(\n  $input: UpdateOrganizationInput!\n) {\n  updateOrganization(input: $input) {\n    organization {\n      id\n      name\n      logoUrl\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "85f76939223253c1d102d05cd74f7538";

export default node;
