/**
 * @generated SignedSource<<a937feef611acb8880f7bf6314f58d07>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SAMLEnforcementPolicy = "OFF" | "OPTIONAL" | "REQUIRED";
export type CreateSAMLConfigurationInput = {
  attributeMappings?: SAMLAttributeMappingsInput | null | undefined;
  autoSignupEnabled: boolean;
  emailDomain: string;
  idpCertificate: string;
  idpEntityId: string;
  idpSsoUrl: string;
  organizationId: string;
};
export type SAMLAttributeMappingsInput = {
  email?: string | null | undefined;
  firstName?: string | null | undefined;
  lastName?: string | null | undefined;
  role?: string | null | undefined;
};
export type NewSAMLConfigurationForm_createMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateSAMLConfigurationInput;
};
export type NewSAMLConfigurationForm_createMutation$data = {
  readonly createSAMLConfiguration: {
    readonly samlConfigurationEdge: {
      readonly node: {
        readonly canDelete: boolean;
        readonly canUpdate: boolean;
        readonly domainVerificationToken: string | null | undefined;
        readonly domainVerifiedAt: string | null | undefined;
        readonly emailDomain: string;
        readonly enforcementPolicy: SAMLEnforcementPolicy;
        readonly id: string;
        readonly testLoginUrl: string;
      };
    };
  } | null | undefined;
};
export type NewSAMLConfigurationForm_createMutation = {
  response: NewSAMLConfigurationForm_createMutation$data;
  variables: NewSAMLConfigurationForm_createMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "concreteType": "SAMLConfigurationEdge",
  "kind": "LinkedField",
  "name": "samlConfigurationEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "SAMLConfiguration",
      "kind": "LinkedField",
      "name": "node",
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
          "name": "emailDomain",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "enforcementPolicy",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "domainVerificationToken",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "domainVerifiedAt",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "testLoginUrl",
          "storageKey": null
        },
        {
          "alias": "canUpdate",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "iam:saml-configuration:update"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"iam:saml-configuration:update\")"
        },
        {
          "alias": "canDelete",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "iam:saml-configuration:delete"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"iam:saml-configuration:delete\")"
        }
      ],
      "storageKey": null
    }
  ],
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "NewSAMLConfigurationForm_createMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateSAMLConfigurationPayload",
        "kind": "LinkedField",
        "name": "createSAMLConfiguration",
        "plural": false,
        "selections": [
          (v3/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "NewSAMLConfigurationForm_createMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateSAMLConfigurationPayload",
        "kind": "LinkedField",
        "name": "createSAMLConfiguration",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "samlConfigurationEdge",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "6ea275ac351f162aeeff7d75d7a3974d",
    "id": null,
    "metadata": {},
    "name": "NewSAMLConfigurationForm_createMutation",
    "operationKind": "mutation",
    "text": "mutation NewSAMLConfigurationForm_createMutation(\n  $input: CreateSAMLConfigurationInput!\n) {\n  createSAMLConfiguration(input: $input) {\n    samlConfigurationEdge {\n      node {\n        id\n        emailDomain\n        enforcementPolicy\n        domainVerificationToken\n        domainVerifiedAt\n        testLoginUrl\n        canUpdate: permission(action: \"iam:saml-configuration:update\")\n        canDelete: permission(action: \"iam:saml-configuration:delete\")\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "af2894da24b742843908a60a7e4629f1";

export default node;
