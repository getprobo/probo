/**
 * @generated SignedSource<<a2a78b5520cedc03e02d257a902e5ee6>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageStatusSectionFragment$data = {
  readonly compliancePage: {
    readonly active: boolean;
    readonly canUpdate: boolean;
    readonly id: string;
  } | null | undefined;
  readonly customDomain: {
    readonly domain: string;
  } | null | undefined;
  readonly " $fragmentType": "CompliancePageStatusSectionFragment";
};
export type CompliancePageStatusSectionFragment$key = {
  readonly " $data"?: CompliancePageStatusSectionFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageStatusSectionFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageStatusSectionFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "CustomDomain",
      "kind": "LinkedField",
      "name": "customDomain",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "domain",
          "storageKey": null
        }
      ],
      "storageKey": null
    },
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
          "name": "active",
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
        }
      ],
      "storageKey": null
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "07af542e3be959de0daf6fcb165830b0";

export default node;
