/**
 * @generated SignedSource<<9e0e298dd1e700d70d3f054b62b43d87>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SAMLEnforcementPolicy = "OFF" | "OPTIONAL" | "REQUIRED";
export type UpdateSAMLConfigurationInput = {
  attributeMappings?: SAMLAttributeMappingsInput | null | undefined;
  autoSignupEnabled?: boolean | null | undefined;
  enforcementPolicy: SAMLEnforcementPolicy;
  idpCertificate?: string | null | undefined;
  idpEntityId?: string | null | undefined;
  idpSsoUrl?: string | null | undefined;
  organizationId: string;
  samlConfigurationId: string;
};
export type SAMLAttributeMappingsInput = {
  email?: string | null | undefined;
  firstName?: string | null | undefined;
  lastName?: string | null | undefined;
  role?: string | null | undefined;
};
export type EditSAMLConfigurationForm_updateMutation$variables = {
  input: UpdateSAMLConfigurationInput;
};
export type EditSAMLConfigurationForm_updateMutation$data = {
  readonly updateSAMLConfiguration: {
    readonly samlConfiguration: {
      readonly domainVerificationToken: string | null | undefined;
      readonly domainVerifiedAt: string | null | undefined;
      readonly emailDomain: string;
      readonly enforcementPolicy: SAMLEnforcementPolicy;
      readonly id: string;
      readonly testLoginUrl: string;
    } | null | undefined;
  } | null | undefined;
};
export type EditSAMLConfigurationForm_updateMutation = {
  response: EditSAMLConfigurationForm_updateMutation$data;
  variables: EditSAMLConfigurationForm_updateMutation$variables;
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
    "concreteType": "UpdateSAMLConfigurationPayload",
    "kind": "LinkedField",
    "name": "updateSAMLConfiguration",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "SAMLConfiguration",
        "kind": "LinkedField",
        "name": "samlConfiguration",
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
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "EditSAMLConfigurationForm_updateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "EditSAMLConfigurationForm_updateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "56d29623294c243a01bfbc2ab4310d28",
    "id": null,
    "metadata": {},
    "name": "EditSAMLConfigurationForm_updateMutation",
    "operationKind": "mutation",
    "text": "mutation EditSAMLConfigurationForm_updateMutation(\n  $input: UpdateSAMLConfigurationInput!\n) {\n  updateSAMLConfiguration(input: $input) {\n    samlConfiguration {\n      id\n      emailDomain\n      enforcementPolicy\n      domainVerificationToken\n      domainVerifiedAt\n      testLoginUrl\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "b98e701514c179bba0f0ce63aae0c0b8";

export default node;
