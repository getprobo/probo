/**
 * @generated SignedSource<<f6ec917cebddcf0a29c1399862127d2b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type SSLStatus = "ACTIVE" | "EXPIRED" | "FAILED" | "PENDING" | "PROVISIONING" | "RENEWING";
import { FragmentRefs } from "relay-runtime";
export type CompliancePageDomainDialogFragment$data = {
  readonly dnsRecords: ReadonlyArray<{
    readonly name: string;
    readonly purpose: string;
    readonly ttl: number;
    readonly type: string;
    readonly value: string;
  }>;
  readonly domain: string;
  readonly sslExpiresAt: string | null | undefined;
  readonly sslStatus: SSLStatus;
  readonly " $fragmentType": "CompliancePageDomainDialogFragment";
};
export type CompliancePageDomainDialogFragment$key = {
  readonly " $data"?: CompliancePageDomainDialogFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageDomainDialogFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageDomainDialogFragment",
  "selections": [
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
      "kind": "ScalarField",
      "name": "domain",
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
      "name": "sslExpiresAt",
      "storageKey": null
    }
  ],
  "type": "CustomDomain",
  "abstractKey": null
};

(node as any).hash = "0275b6b52781473cb3dac5a631488e2f";

export default node;
