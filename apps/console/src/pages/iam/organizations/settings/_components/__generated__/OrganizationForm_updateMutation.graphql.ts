/**
 * @generated SignedSource<<9602c1a0ad55d4889431c8e5f54f4461>>
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
export type OrganizationForm_updateMutation$variables = {
  input: UpdateOrganizationInput;
};
export type OrganizationForm_updateMutation$data = {
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
    } | null | undefined;
  } | null | undefined;
};
export type OrganizationForm_updateMutation = {
  response: OrganizationForm_updateMutation$data;
  variables: OrganizationForm_updateMutation$variables;
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
    "name": "OrganizationForm_updateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "OrganizationForm_updateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "f2a249d13ff38645cb858b494c05c59d",
    "id": null,
    "metadata": {},
    "name": "OrganizationForm_updateMutation",
    "operationKind": "mutation",
    "text": "mutation OrganizationForm_updateMutation(\n  $input: UpdateOrganizationInput!\n) {\n  updateOrganization(input: $input) {\n    organization {\n      id\n      name\n      logoUrl\n      horizontalLogoUrl\n      description\n      websiteUrl\n      email\n      headquarterAddress\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "c4dd7a42f611fa210594017a5a63d159";

export default node;
