/**
 * @generated SignedSource<<92455d5992a9e72cb7d395f570ad308c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type VendorOverviewTabBusinessAssociateAgreementFragment$data = {
  readonly businessAssociateAgreement: {
    readonly canDelete: boolean;
    readonly canUpdate: boolean;
    readonly createdAt: string;
    readonly fileName: string;
    readonly fileUrl: string;
    readonly id: string;
    readonly validFrom: string | null | undefined;
    readonly validUntil: string | null | undefined;
  } | null | undefined;
  readonly " $fragmentType": "VendorOverviewTabBusinessAssociateAgreementFragment";
};
export type VendorOverviewTabBusinessAssociateAgreementFragment$key = {
  readonly " $data"?: VendorOverviewTabBusinessAssociateAgreementFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"VendorOverviewTabBusinessAssociateAgreementFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "VendorOverviewTabBusinessAssociateAgreementFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "VendorBusinessAssociateAgreement",
      "kind": "LinkedField",
      "name": "businessAssociateAgreement",
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
          "name": "fileName",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "fileUrl",
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
          "name": "createdAt",
          "storageKey": null
        },
        {
          "alias": "canUpdate",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:vendor-business-associate-agreement:update"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:vendor-business-associate-agreement:update\")"
        },
        {
          "alias": "canDelete",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:vendor-business-associate-agreement:delete"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:vendor-business-associate-agreement:delete\")"
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Vendor",
  "abstractKey": null
};

(node as any).hash = "3c3bb156f86684d1834ec63cf43dc99f";

export default node;
