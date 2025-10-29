/**
 * @generated SignedSource<<694befcb248cbb23e66b9555626f223b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type SAMLEnforcementPolicy = "OFF" | "OPTIONAL" | "REQUIRED";
export type UpdateSAMLConfigurationInput = {
  attributeEmail?: string | null | undefined;
  attributeFirstname?: string | null | undefined;
  attributeLastname?: string | null | undefined;
  attributeRole?: string | null | undefined;
  autoSignupEnabled?: boolean | null | undefined;
  defaultRole?: string | null | undefined;
  enabled?: boolean | null | undefined;
  enforcementPolicy?: SAMLEnforcementPolicy | null | undefined;
  id: string;
  idpCertificate?: string | null | undefined;
  idpEntityId?: string | null | undefined;
  idpMetadataUrl?: string | null | undefined;
  idpSsoUrl?: string | null | undefined;
  spCertificate?: string | null | undefined;
  spPrivateKey?: string | null | undefined;
};
export type SAMLConfigurationGraphUpdateMutation$variables = {
  input: UpdateSAMLConfigurationInput;
};
export type SAMLConfigurationGraphUpdateMutation$data = {
  readonly updateSAMLConfiguration: {
    readonly samlConfiguration: {
      readonly attributeEmail: string;
      readonly attributeFirstname: string;
      readonly attributeLastname: string;
      readonly attributeRole: string;
      readonly autoSignupEnabled: boolean;
      readonly createdAt: any;
      readonly defaultRole: string;
      readonly emailDomain: string;
      readonly enabled: boolean;
      readonly enforcementPolicy: SAMLEnforcementPolicy;
      readonly id: string;
      readonly idpCertificate: string;
      readonly idpEntityId: string;
      readonly idpMetadataUrl: string | null | undefined;
      readonly idpSsoUrl: string;
      readonly spAcsUrl: string;
      readonly spEntityId: string;
      readonly spMetadataUrl: string;
      readonly testLoginUrl: string;
      readonly updatedAt: any;
    };
  };
};
export type SAMLConfigurationGraphUpdateMutation = {
  response: SAMLConfigurationGraphUpdateMutation$data;
  variables: SAMLConfigurationGraphUpdateMutation$variables;
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
            "name": "enabled",
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
            "name": "spEntityId",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "spAcsUrl",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "spMetadataUrl",
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
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "idpEntityId",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "idpSsoUrl",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "idpCertificate",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "idpMetadataUrl",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "attributeEmail",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "attributeFirstname",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "attributeLastname",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "attributeRole",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "defaultRole",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "autoSignupEnabled",
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
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "SAMLConfigurationGraphUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SAMLConfigurationGraphUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "c9e4055888055b109aa71b665c9544ca",
    "id": null,
    "metadata": {},
    "name": "SAMLConfigurationGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation SAMLConfigurationGraphUpdateMutation(\n  $input: UpdateSAMLConfigurationInput!\n) {\n  updateSAMLConfiguration(input: $input) {\n    samlConfiguration {\n      id\n      enabled\n      emailDomain\n      enforcementPolicy\n      spEntityId\n      spAcsUrl\n      spMetadataUrl\n      testLoginUrl\n      idpEntityId\n      idpSsoUrl\n      idpCertificate\n      idpMetadataUrl\n      attributeEmail\n      attributeFirstname\n      attributeLastname\n      attributeRole\n      defaultRole\n      autoSignupEnabled\n      createdAt\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "d572af05811d6a05cdd4446fe175894d";

export default node;
