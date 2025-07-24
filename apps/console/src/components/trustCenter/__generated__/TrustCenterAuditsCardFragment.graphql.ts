/**
 * @generated SignedSource<<794277d17daabd92807ccfaa5a1c2dbb>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type AuditState = "COMPLETED" | "IN_PROGRESS" | "NOT_STARTED" | "OUTDATED" | "REJECTED";
import { FragmentRefs } from "relay-runtime";
export type TrustCenterAuditsCardFragment$data = {
  readonly createdAt: any;
  readonly framework: {
    readonly name: string;
  };
  readonly id: string;
  readonly showOnTrustCenter: boolean;
  readonly state: AuditState;
  readonly validFrom: any | null | undefined;
  readonly validUntil: any | null | undefined;
  readonly " $fragmentType": "TrustCenterAuditsCardFragment";
};
export type TrustCenterAuditsCardFragment$key = {
  readonly " $data"?: TrustCenterAuditsCardFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"TrustCenterAuditsCardFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "TrustCenterAuditsCardFragment",
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
      "concreteType": "Framework",
      "kind": "LinkedField",
      "name": "framework",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "name",
          "storageKey": null
        }
      ],
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "validFrom",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "validUntil",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "state",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "showOnTrustCenter",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "createdAt",
      "storageKey": null
    }
  ],
  "type": "Audit",
  "abstractKey": null
};

(node as any).hash = "9c9291dffd5c057c0e70e2567a897362";

export default node;
