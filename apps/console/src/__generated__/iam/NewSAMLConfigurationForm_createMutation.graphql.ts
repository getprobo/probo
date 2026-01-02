/**
 * @generated SignedSource<<f88cc717a937f85bb376bd359d7ca722>>
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
    "cacheID": "969d5fb34995f5cce359c792cbe1ff49",
    "id": null,
    "metadata": {},
    "name": "NewSAMLConfigurationForm_createMutation",
    "operationKind": "mutation",
    "text": "mutation NewSAMLConfigurationForm_createMutation(\n  $input: CreateSAMLConfigurationInput!\n) {\n  createSAMLConfiguration(input: $input) {\n    samlConfigurationEdge {\n      node {\n        id\n        emailDomain\n        enforcementPolicy\n        domainVerificationToken\n        domainVerifiedAt\n        testLoginUrl\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "5a3b2e5d219b40ca096eece75d7b63b3";

export default node;
