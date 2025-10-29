/**
 * @generated SignedSource<<8e0818a9214ba3613f9d356e61a75e08>>
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
export type GeneralSettingsTab_DeleteHorizontalLogoMutation$variables = {
  input: DeleteOrganizationHorizontalLogoInput;
};
export type GeneralSettingsTab_DeleteHorizontalLogoMutation$data = {
  readonly deleteOrganizationHorizontalLogo: {
    readonly organization: {
      readonly horizontalLogoUrl: string | null | undefined;
      readonly id: string;
    };
  };
};
export type GeneralSettingsTab_DeleteHorizontalLogoMutation = {
  response: GeneralSettingsTab_DeleteHorizontalLogoMutation$data;
  variables: GeneralSettingsTab_DeleteHorizontalLogoMutation$variables;
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
    "name": "GeneralSettingsTab_DeleteHorizontalLogoMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "GeneralSettingsTab_DeleteHorizontalLogoMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "fbfaf507b48e2ef274e44372a70b3d88",
    "id": null,
    "metadata": {},
    "name": "GeneralSettingsTab_DeleteHorizontalLogoMutation",
    "operationKind": "mutation",
    "text": "mutation GeneralSettingsTab_DeleteHorizontalLogoMutation(\n  $input: DeleteOrganizationHorizontalLogoInput!\n) {\n  deleteOrganizationHorizontalLogo(input: $input) {\n    organization {\n      id\n      horizontalLogoUrl\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "7910936d423f99e36ee0a082f5c9336c";

export default node;
