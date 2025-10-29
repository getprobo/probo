/**
 * @generated SignedSource<<ad8f1e5fd866cefc124b7b7c78bce367>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type SAMLEnforcementPolicy = "OFF" | "OPTIONAL" | "REQUIRED";
import { FragmentRefs } from "relay-runtime";
export type SAMLSettingsTabFragment$data = {
  readonly id: string;
  readonly name: string;
  readonly samlConfigurations: ReadonlyArray<{
    readonly attributeEmail: string;
    readonly attributeFirstname: string;
    readonly attributeLastname: string;
    readonly attributeRole: string;
    readonly autoSignupEnabled: boolean;
    readonly domainVerificationToken: string | null | undefined;
    readonly domainVerified: boolean;
    readonly domainVerifiedAt: any | null | undefined;
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
  }>;
  readonly " $fragmentType": "SAMLSettingsTabFragment";
};
export type SAMLSettingsTabFragment$key = {
  readonly " $data"?: SAMLSettingsTabFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"SAMLSettingsTabFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "SAMLSettingsTabFragment",
  "selections": [
    (v0/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "name",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "SAMLConfiguration",
      "kind": "LinkedField",
      "name": "samlConfigurations",
      "plural": true,
      "selections": [
        (v0/*: any*/),
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
          "name": "domainVerified",
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
          "name": "autoSignupEnabled",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Organization",
  "abstractKey": null
};
})();

(node as any).hash = "eec380838cfd22d70a504028bb5046b4";

export default node;
