/**
 * @generated SignedSource<<cf0ce3993d747999dfbe2ca60b509023>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type SSLStatus = "ACTIVE" | "EXPIRED" | "FAILED" | "PENDING" | "PROVISIONING" | "RENEWING";
import { FragmentRefs } from "relay-runtime";
export type SettingsPageFragment$data = {
  readonly createdAt: any;
  readonly customDomain: {
    readonly createdAt: any;
    readonly dnsRecords: ReadonlyArray<{
      readonly name: string;
      readonly purpose: string;
      readonly ttl: number;
      readonly type: string;
      readonly value: string;
    }>;
    readonly domain: string;
    readonly id: string;
    readonly sslExpiresAt: any | null | undefined;
    readonly sslStatus: SSLStatus;
    readonly updatedAt: any;
  } | null | undefined;
  readonly description: string | null | undefined;
  readonly email: string | null | undefined;
  readonly headquarterAddress: string | null | undefined;
  readonly horizontalLogoUrl: string | null | undefined;
  readonly id: string;
  readonly logoUrl: string | null | undefined;
  readonly name: string;
  readonly updatedAt: any;
  readonly websiteUrl: string | null | undefined;
  readonly " $fragmentType": "SettingsPageFragment";
};
export type SettingsPageFragment$key = {
  readonly " $data"?: SettingsPageFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"SettingsPageFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "SettingsPageFragment",
  "selections": [
    (v0/*: any*/),
    (v1/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "logoUrl",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "horizontalLogoUrl",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "description",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "websiteUrl",
      "storageKey": null
    },
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
      "name": "headquarterAddress",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "CustomDomain",
      "kind": "LinkedField",
      "name": "customDomain",
      "plural": false,
      "selections": [
        (v0/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "domain",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "sslStatus",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "concreteType": "DNSRecordInstruction",
          "kind": "LinkedField",
          "name": "dnsRecords",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "type",
              "storageKey": null
            },
            (v1/*: any*/),
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "value",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "ttl",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "purpose",
              "storageKey": null
            }
          ],
          "storageKey": null
        },
        (v2/*: any*/),
        (v3/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "sslExpiresAt",
          "storageKey": null
        }
      ],
      "storageKey": null
    },
    (v2/*: any*/),
    (v3/*: any*/)
  ],
  "type": "Organization",
  "abstractKey": null
};
})();

(node as any).hash = "1e64bcccf8ef3a8ead79b3446e8a3ccd";

export default node;
