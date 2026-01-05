/**
 * @generated SignedSource<<1d007308b8b70962ddfe217cf3534dab>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type SSLStatus = "ACTIVE" | "EXPIRED" | "FAILED" | "PENDING" | "PROVISIONING" | "RENEWING";
import { FragmentRefs } from "relay-runtime";
export type DomainDialogFragment$data = {
  readonly createdAt: string;
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
  readonly updatedAt: string;
  readonly " $fragmentType": "DomainDialogFragment";
};
export type DomainDialogFragment$key = {
  readonly " $data"?: DomainDialogFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DomainDialogFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DomainDialogFragment",
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
  "type": "CustomDomain",
  "abstractKey": null
};

(node as any).hash = "0aec5615023dde2d23901fec9e43ed9b";

export default node;
