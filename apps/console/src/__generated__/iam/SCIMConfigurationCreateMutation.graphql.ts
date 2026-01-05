/**
 * @generated SignedSource<<5a8bd03c48696bd6ed7bd5ec80a61045>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateSCIMConfigurationInput = {
  organizationId: string;
};
export type SCIMConfigurationCreateMutation$variables = {
  input: CreateSCIMConfigurationInput;
};
export type SCIMConfigurationCreateMutation$data = {
  readonly createSCIMConfiguration: {
    readonly scimConfiguration: {
      readonly organization: {
        readonly id: string;
        readonly scimConfiguration: {
          readonly createdAt: string;
          readonly endpointUrl: string;
          readonly id: string;
          readonly updatedAt: string;
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
v1 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "concreteType": "Organization",
  "kind": "LinkedField",
  "name": "organization",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    {
      "alias": null,
      "args": null,
      "concreteType": "SCIMConfiguration",
      "kind": "LinkedField",
      "name": "scimConfiguration",
      "plural": false,
      "selections": [
        (v2/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "endpointUrl",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "createdAt",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "updatedAt",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "token",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "SCIMConfigurationCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
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
              (v3/*: any*/)
            ],
            "storageKey": null
          },
          (v4/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SCIMConfigurationCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
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
              (v3/*: any*/),
              (v2/*: any*/)
            ],
            "storageKey": null
          },
          (v4/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "24fe7551556768ace60e5c1ee357deda",
    "id": null,
    "metadata": {},
    "name": "SCIMConfigurationCreateMutation",
    "operationKind": "mutation",
    "text": "mutation SCIMConfigurationCreateMutation(\n  $input: CreateSCIMConfigurationInput!\n) {\n  createSCIMConfiguration(input: $input) {\n    scimConfiguration {\n      organization {\n        id\n        scimConfiguration {\n          id\n          endpointUrl\n          createdAt\n          updatedAt\n        }\n      }\n      id\n    }\n    token\n  }\n}\n"
  }
};
})();

(node as any).hash = "3b23015e5a2e1b117317cbfad4e4a64b";

export default node;
