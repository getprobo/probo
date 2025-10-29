/**
 * @generated SignedSource<<270832e99647c6636914a9644a65358c>>
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
  horizontalLogoFile?: any | null | undefined;
  logoFile?: any | null | undefined;
  name?: string | null | undefined;
  organizationId: string;
  websiteUrl?: string | null | undefined;
};
export type GeneralSettingsTab_UpdateMutation$variables = {
  input: UpdateOrganizationInput;
};
export type GeneralSettingsTab_UpdateMutation$data = {
  readonly updateOrganization: {
    readonly organization: {
      readonly description: string | null | undefined;
      readonly email: string | null | undefined;
      readonly headquarterAddress: string | null | undefined;
      readonly horizontalLogoUrl: string | null | undefined;
      readonly id: string;
      readonly logoUrl: string | null | undefined;
      readonly name: string;
      readonly websiteUrl: string | null | undefined;
    };
  };
};
export type GeneralSettingsTab_UpdateMutation = {
  response: GeneralSettingsTab_UpdateMutation$data;
  variables: GeneralSettingsTab_UpdateMutation$variables;
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
            "name": "horizontalLogoUrl",
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
    "name": "GeneralSettingsTab_UpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "GeneralSettingsTab_UpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "55c07b334317a5ca30023ef5354a6c45",
    "id": null,
    "metadata": {},
    "name": "GeneralSettingsTab_UpdateMutation",
    "operationKind": "mutation",
    "text": "mutation GeneralSettingsTab_UpdateMutation(\n  $input: UpdateOrganizationInput!\n) {\n  updateOrganization(input: $input) {\n    organization {\n      id\n      name\n      logoUrl\n      horizontalLogoUrl\n      description\n      websiteUrl\n      email\n      headquarterAddress\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "f1731adcb4bf7b7214301a612f48567e";

export default node;
