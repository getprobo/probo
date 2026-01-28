/**
 * @generated SignedSource<<b25caa52d25189aca17270cef3c0303e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageSlackSectionFragment$data = {
  readonly compliancePage: {
    readonly canUpdate: boolean;
  } | null | undefined;
  readonly slackConnections: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly channel: string | null | undefined;
        readonly createdAt: string;
        readonly id: string;
      };
    }>;
  };
  readonly " $fragmentType": "CompliancePageSlackSectionFragment";
};
export type CompliancePageSlackSectionFragment$key = {
  readonly " $data"?: CompliancePageSlackSectionFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageSlackSectionFragment">;
};

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageSlackSectionFragment",
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
    },
    {
      "alias": null,
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 100
        }
      ],
      "concreteType": "SlackConnectionConnection",
      "kind": "LinkedField",
      "name": "slackConnections",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "SlackConnectionEdge",
          "kind": "LinkedField",
          "name": "edges",
          "plural": true,
          "selections": [
            {
              "alias": null,
              "args": null,
              "concreteType": "SlackConnection",
              "kind": "LinkedField",
              "name": "node",
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
                  "name": "channel",
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
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": "slackConnections(first:100)"
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "48af0130884a7e0070e64ead2afe5046";

export default node;
