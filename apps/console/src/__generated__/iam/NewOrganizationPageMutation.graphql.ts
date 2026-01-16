/**
 * @generated SignedSource<<3c4bf3ac3f3eff32d0c86fe61190576e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateOrganizationInput = {
  horizontalLogoFile?: any | null | undefined;
  logoFile?: any | null | undefined;
  name: string;
};
export type NewOrganizationPageMutation$variables = {
  input: CreateOrganizationInput;
};
export type NewOrganizationPageMutation$data = {
  readonly createOrganization: {
    readonly organization: {
      readonly id: string;
      readonly name: string;
    } | null | undefined;
  } | null | undefined;
};
export type NewOrganizationPageMutation = {
  response: NewOrganizationPageMutation$data;
  variables: NewOrganizationPageMutation$variables;
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
    "concreteType": "CreateOrganizationPayload",
    "kind": "LinkedField",
    "name": "createOrganization",
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
    "name": "NewOrganizationPageMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "NewOrganizationPageMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "e4d07eba4c1791eeb6e5e1e627c79593",
    "id": null,
    "metadata": {},
    "name": "NewOrganizationPageMutation",
    "operationKind": "mutation",
    "text": "mutation NewOrganizationPageMutation(\n  $input: CreateOrganizationInput!\n) {\n  createOrganization(input: $input) {\n    organization {\n      id\n      name\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "68737a40f2993357fb169e9aac6a4262";

export default node;
