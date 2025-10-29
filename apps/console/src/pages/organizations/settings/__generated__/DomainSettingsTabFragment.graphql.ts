/**
 * @generated SignedSource<<0e2aa976b1c9bb8dddf8b1dcc2f171d1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type SSLStatus = "ACTIVE" | "EXPIRED" | "FAILED" | "PENDING" | "PROVISIONING" | "RENEWING";
import { FragmentRefs } from "relay-runtime";
export type DomainSettingsTabFragment$data = {
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
  readonly id: string;
  readonly " $fragmentType": "DomainSettingsTabFragment";
};
export type DomainSettingsTabFragment$key = {
  readonly " $data"?: DomainSettingsTabFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DomainSettingsTabFragment">;
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
  "name": "DomainSettingsTabFragment",
  "selections": [
    (v0/*: any*/),
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
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "sslExpiresAt",
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

(node as any).hash = "00306efb96d302284155f5324ba2fb99";

export default node;
