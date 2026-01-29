/**
 * @generated SignedSource<<29e2f3791be9e59a1c61033e31eb0134>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateSCIMConfigurationInput = {
  connectorId?: string | null | undefined;
  organizationId: string;
};
export type SCIMConfigurationCreateMutation$variables = {
  input: CreateSCIMConfigurationInput;
};
export type SCIMConfigurationCreateMutation$data = {
  readonly createSCIMConfiguration: {
    readonly scimConfiguration: {
      readonly endpointUrl: string;
      readonly id: string;
      readonly organization: {
        readonly id: string;
        readonly scimConfiguration: {
          readonly endpointUrl: string;
          readonly id: string;
        } | null | undefined;
      } | null | undefined;
    };
    readonly token: string;
  } | null | undefined;
};
export type SCIMConfigurationCreateMutation = {
  response: SCIMConfigurationCreateMutation$data;
  variables: SCIMConfigurationCreateMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "endpointUrl",
  "storageKey": null
},
v3 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "CreateSCIMConfigurationPayload",
    "kind": "LinkedField",
    "name": "createSCIMConfiguration",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "SCIMConfiguration",
        "kind": "LinkedField",
        "name": "scimConfiguration",
        "plural": false,
        "selections": [
          (v1/*: any*/),
          (v2/*: any*/),
          {
            "alias": null,
            "args": null,
            "concreteType": "Organization",
            "kind": "LinkedField",
            "name": "organization",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "SCIMConfiguration",
                "kind": "LinkedField",
                "name": "scimConfiguration",
                "plural": false,
                "selections": [
                  (v1/*: any*/),
                  (v2/*: any*/)
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "token",
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
    "name": "SCIMConfigurationCreateMutation",
    "selections": (v3/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SCIMConfigurationCreateMutation",
    "selections": (v3/*: any*/)
  },
  "params": {
    "cacheID": "c4a8c2c34586e55c6dca0b5e4a0f4642",
    "id": null,
    "metadata": {},
    "name": "SCIMConfigurationCreateMutation",
    "operationKind": "mutation",
    "text": "mutation SCIMConfigurationCreateMutation(\n  $input: CreateSCIMConfigurationInput!\n) {\n  createSCIMConfiguration(input: $input) {\n    scimConfiguration {\n      id\n      endpointUrl\n      organization {\n        id\n        scimConfiguration {\n          id\n          endpointUrl\n        }\n      }\n    }\n    token\n  }\n}\n"
  }
};
})();

(node as any).hash = "b4534941bf2670d13c78cb5ab9e810b1";

export default node;
