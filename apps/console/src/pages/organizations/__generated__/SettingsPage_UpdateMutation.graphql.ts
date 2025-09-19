/**
 * @generated SignedSource<<892a78db951d547b53d595d219b94e75>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateOrganizationInput = {
  description?: string | null | undefined;
  email?: string | null | undefined;
  headquarterAddress?: string | null | undefined;
  logo?: any | null | undefined;
  name?: string | null | undefined;
  organizationId: string;
  websiteUrl?: string | null | undefined;
};
export type SettingsPage_UpdateMutation$variables = {
  input: UpdateOrganizationInput;
};
export type SettingsPage_UpdateMutation$data = {
  readonly updateOrganization: {
    readonly organization: {
      readonly description: string | null | undefined;
      readonly email: string | null | undefined;
      readonly headquarterAddress: string | null | undefined;
      readonly id: string;
      readonly logoUrl: string | null | undefined;
      readonly name: string;
      readonly websiteUrl: string | null | undefined;
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
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "description",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "websiteUrl",
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
            "name": "headquarterAddress",
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
    "cacheID": "3526d270f311db33e0563abc99c96978",
    "id": null,
    "metadata": {},
    "name": "SettingsPage_UpdateMutation",
    "operationKind": "mutation",
    "text": "mutation SettingsPage_UpdateMutation(\n  $input: UpdateOrganizationInput!\n) {\n  updateOrganization(input: $input) {\n    organization {\n      id\n      name\n      logoUrl\n      description\n      websiteUrl\n      email\n      headquarterAddress\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "54949158defa7f1bb8e1e578db7cd7be";

export default node;
