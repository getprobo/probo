/**
 * @generated SignedSource<<f34032526d9b51e3b675aad1041d3796>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageNDASectionFragment$data = {
  readonly compliancePage: {
    readonly canDeleteNDA: boolean;
    readonly canUpdate: boolean;
    readonly canUploadNDA: boolean;
    readonly id: string;
    readonly ndaFileName: string | null | undefined;
    readonly ndaFileUrl: string | null | undefined;
  } | null | undefined;
  readonly " $fragmentType": "CompliancePageNDASectionFragment";
};
export type CompliancePageNDASectionFragment$key = {
  readonly " $data"?: CompliancePageNDASectionFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageNDASectionFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageNDASectionFragment",
  "selections": [
    {
      "alias": "compliancePage",
      "args": null,
      "concreteType": "TrustCenter",
      "kind": "LinkedField",
      "name": "trustCenter",
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
          "name": "ndaFileName",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "ndaFileUrl",
          "storageKey": null
        },
        {
          "alias": "canUpdate",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:trust-center:update"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:trust-center:update\")"
        },
        {
          "alias": "canUploadNDA",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:trust-center:upload-nda"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:trust-center:upload-nda\")"
        },
        {
          "alias": "canDeleteNDA",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:trust-center:delete-nda"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:trust-center:delete-nda\")"
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "3a55696877cb2d258edf0b9ed615de91";

export default node;
