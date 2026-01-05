/**
 * @generated SignedSource<<9f5f726789193f7af37d789c8b324a1e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SAMLEnforcementPolicy = "OFF" | "OPTIONAL" | "REQUIRED";
export type EditSAMLConfigurationFormQuery$variables = {
  samlConfigurationId: string;
};
export type EditSAMLConfigurationFormQuery$data = {
  readonly samlConfiguration: {
    readonly __typename: "SAMLConfiguration";
    readonly attributeMappings: {
      readonly email: string;
      readonly firstName: string;
      readonly lastName: string;
      readonly role: string;
    };
    readonly autoSignupEnabled: boolean;
    readonly domainVerificationToken: string | null | undefined;
    readonly domainVerifiedAt: string | null | undefined;
    readonly emailDomain: string;
    readonly enforcementPolicy: SAMLEnforcementPolicy;
    readonly id: string;
    readonly idpCertificate: string;
    readonly idpEntityId: string;
    readonly idpSsoUrl: string;
    readonly testLoginUrl: string;
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
};
export type EditSAMLConfigurationFormQuery = {
  response: EditSAMLConfigurationFormQuery$data;
  variables: EditSAMLConfigurationFormQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "samlConfigurationId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "samlConfigurationId"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "emailDomain",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "enforcementPolicy",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "domainVerificationToken",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "domainVerifiedAt",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "testLoginUrl",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "idpEntityId",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "idpSsoUrl",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "idpCertificate",
  "storageKey": null
},
v12 = {
  "alias": null,
  "args": null,
  "concreteType": "SAMLAttributeMappings",
  "kind": "LinkedField",
  "name": "attributeMappings",
  "plural": false,
  "selections": [
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
      "name": "firstName",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "lastName",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "role",
      "storageKey": null
    }
  ],
  "storageKey": null
},
v13 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "autoSignupEnabled",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "EditSAMLConfigurationFormQuery",
    "selections": [
      {
        "kind": "RequiredField",
        "field": {
          "alias": "samlConfiguration",
          "args": (v1/*: any*/),
          "concreteType": null,
          "kind": "LinkedField",
          "name": "node",
          "plural": false,
          "selections": [
            (v2/*: any*/),
            {
              "kind": "InlineFragment",
              "selections": [
                (v3/*: any*/),
                (v4/*: any*/),
                (v5/*: any*/),
                (v6/*: any*/),
                (v7/*: any*/),
                (v8/*: any*/),
                (v9/*: any*/),
                (v10/*: any*/),
                (v11/*: any*/),
                (v12/*: any*/),
                (v13/*: any*/)
              ],
              "type": "SAMLConfiguration",
              "abstractKey": null
            }
          ],
          "storageKey": null
        },
        "action": "THROW"
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "EditSAMLConfigurationFormQuery",
    "selections": [
      {
        "alias": "samlConfiguration",
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          (v3/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v12/*: any*/),
              (v13/*: any*/)
            ],
            "type": "SAMLConfiguration",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "47c1ce92e8ba0395fa803b343d093879",
    "id": null,
    "metadata": {},
    "name": "EditSAMLConfigurationFormQuery",
    "operationKind": "query",
    "text": "query EditSAMLConfigurationFormQuery(\n  $samlConfigurationId: ID!\n) {\n  samlConfiguration: node(id: $samlConfigurationId) {\n    __typename\n    ... on SAMLConfiguration {\n      id\n      emailDomain\n      enforcementPolicy\n      domainVerificationToken\n      domainVerifiedAt\n      testLoginUrl\n      idpEntityId\n      idpSsoUrl\n      idpCertificate\n      attributeMappings {\n        email\n        firstName\n        lastName\n        role\n      }\n      autoSignupEnabled\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "d31326e3e4423e62d8acde5735d9b578";

export default node;
