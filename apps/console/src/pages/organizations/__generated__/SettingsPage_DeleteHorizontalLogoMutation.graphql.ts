/**
 * @generated SignedSource<<4ed3d530d746d8b84b2dba8752e57abc>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteOrganizationHorizontalLogoInput = {
  organizationId: string;
};
export type SettingsPage_DeleteHorizontalLogoMutation$variables = {
  input: DeleteOrganizationHorizontalLogoInput;
};
export type SettingsPage_DeleteHorizontalLogoMutation$data = {
  readonly deleteOrganizationHorizontalLogo: {
    readonly organization: {
      readonly horizontalLogoUrl: string | null | undefined;
      readonly id: string;
    };
  };
};
export type SettingsPage_DeleteHorizontalLogoMutation = {
  response: SettingsPage_DeleteHorizontalLogoMutation$data;
  variables: SettingsPage_DeleteHorizontalLogoMutation$variables;
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
    "concreteType": "DeleteOrganizationHorizontalLogoPayload",
    "kind": "LinkedField",
    "name": "deleteOrganizationHorizontalLogo",
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
            "name": "horizontalLogoUrl",
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
    "name": "SettingsPage_DeleteHorizontalLogoMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SettingsPage_DeleteHorizontalLogoMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "e631480d50c9347050fdc62075a2e3a3",
    "id": null,
    "metadata": {},
    "name": "SettingsPage_DeleteHorizontalLogoMutation",
    "operationKind": "mutation",
    "text": "mutation SettingsPage_DeleteHorizontalLogoMutation(\n  $input: DeleteOrganizationHorizontalLogoInput!\n) {\n  deleteOrganizationHorizontalLogo(input: $input) {\n    organization {\n      id\n      horizontalLogoUrl\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "751c3ff44c59511451095ffc66446c2f";

export default node;
