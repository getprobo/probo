/**
 * @generated SignedSource<<ea00a00b77edf3939a310de2b0f7b8e0>>
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
export type OrganizationForm_deleteHorizontalLogoMutation$variables = {
  input: DeleteOrganizationHorizontalLogoInput;
};
export type OrganizationForm_deleteHorizontalLogoMutation$data = {
  readonly deleteOrganizationHorizontalLogo: {
    readonly organization: {
      readonly horizontalLogoUrl: string | null | undefined;
      readonly id: string;
    };
  } | null | undefined;
};
export type OrganizationForm_deleteHorizontalLogoMutation = {
  response: OrganizationForm_deleteHorizontalLogoMutation$data;
  variables: OrganizationForm_deleteHorizontalLogoMutation$variables;
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
    "name": "OrganizationForm_deleteHorizontalLogoMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "OrganizationForm_deleteHorizontalLogoMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "2ec98faaaf8317f642f5db54c7bbed80",
    "id": null,
    "metadata": {},
    "name": "OrganizationForm_deleteHorizontalLogoMutation",
    "operationKind": "mutation",
    "text": "mutation OrganizationForm_deleteHorizontalLogoMutation(\n  $input: DeleteOrganizationHorizontalLogoInput!\n) {\n  deleteOrganizationHorizontalLogo(input: $input) {\n    organization {\n      id\n      horizontalLogoUrl\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "9ea67e345e506fef9d85ee31bc360d51";

export default node;
