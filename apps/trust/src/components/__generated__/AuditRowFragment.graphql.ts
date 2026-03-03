/**
 * @generated SignedSource<<beb9130a1b213333ae647116467056a0>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type DocumentAccessStatus = "GRANTED" | "REJECTED" | "REQUESTED" | "REVOKED";
import { FragmentRefs } from "relay-runtime";
export type AuditRowFragment$data = {
  readonly framework: {
    readonly darkLogoURL: string | null | undefined;
    readonly id: string;
    readonly lightLogoURL: string | null | undefined;
    readonly name: string;
  };
  readonly name: string | null | undefined;
  readonly report: {
    readonly access: {
      readonly id: string;
      readonly status: DocumentAccessStatus;
    } | null | undefined;
    readonly filename: string;
    readonly id: string;
    readonly isUserAuthorized: boolean;
  } | null | undefined;
  readonly " $fragmentType": "AuditRowFragment";
};
export type AuditRowFragment$key = {
  readonly " $data"?: AuditRowFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"AuditRowFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v1 = {
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
  "name": "AuditRowFragment",
  "selections": [
    (v0/*: any*/),
    {
      "alias": null,
      "args": null,
      "concreteType": "Report",
      "kind": "LinkedField",
      "name": "report",
      "plural": false,
      "selections": [
        (v1/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "filename",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "isUserAuthorized",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "concreteType": "DocumentAccess",
          "kind": "LinkedField",
          "name": "access",
          "plural": false,
          "selections": [
            (v1/*: any*/),
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "status",
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
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
        (v1/*: any*/),
        (v0/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "lightLogoURL",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "darkLogoURL",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Audit",
  "abstractKey": null
};
})();

(node as any).hash = "1e3e89368594b05e0926b5ee8bc7c0b5";

export default node;
