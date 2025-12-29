/**
 * @generated SignedSource<<efdbeab080fabb165a3a169fbedf3d8f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type SSLStatus = "ACTIVE" | "EXPIRED" | "FAILED" | "PENDING" | "PROVISIONING" | "RENEWING";
import { FragmentRefs } from "relay-runtime";
export type DomainCardFragment$data = {
  readonly domain: string;
  readonly sslStatus: SSLStatus;
  readonly " $fragmentSpreads": FragmentRefs<"DomainDialogFragment">;
  readonly " $fragmentType": "DomainCardFragment";
};
export type DomainCardFragment$key = {
  readonly " $data"?: DomainCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"DomainCardFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "DomainCardFragment",
  "selections": [
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
      "args": null,
      "kind": "FragmentSpread",
      "name": "DomainDialogFragment"
    }
  ],
  "type": "CustomDomain",
  "abstractKey": null
};

(node as any).hash = "672098f5351cf031d958263155e180ac";

export default node;
