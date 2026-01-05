/**
 * @generated SignedSource<<c1f987833f73f19a66f044b28688f1dd>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CreateControlAuditMappingInput = {
  auditId: string;
  controlId: string;
};
export type FrameworkControlPageAttachAuditMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateControlAuditMappingInput;
};
export type FrameworkControlPageAttachAuditMutation$data = {
  readonly createControlAuditMapping: {
    readonly auditEdge: {
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"LinkedAuditsCardFragment">;
      };
    };
  };
};
export type FrameworkControlPageAttachAuditMutation = {
  response: FrameworkControlPageAttachAuditMutation$data;
  variables: FrameworkControlPageAttachAuditMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "FrameworkControlPageAttachAuditMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateControlAuditMappingPayload",
        "kind": "LinkedField",
        "name": "createControlAuditMapping",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "AuditEdge",
            "kind": "LinkedField",
            "name": "auditEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Audit",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "LinkedAuditsCardFragment"
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "FrameworkControlPageAttachAuditMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateControlAuditMappingPayload",
        "kind": "LinkedField",
        "name": "createControlAuditMapping",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "AuditEdge",
            "kind": "LinkedField",
            "name": "auditEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Audit",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v4/*: any*/),
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
                    "name": "state",
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
                    "concreteType": "Framework",
                    "kind": "LinkedField",
                    "name": "framework",
                    "plural": false,
                    "selections": [
                      (v3/*: any*/),
                      (v4/*: any*/)
                    ],
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
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "auditEdge",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "23afa68a91d9ae6630123a4ff693a0ec",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlPageAttachAuditMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlPageAttachAuditMutation(\n  $input: CreateControlAuditMappingInput!\n) {\n  createControlAuditMapping(input: $input) {\n    auditEdge {\n      node {\n        id\n        ...LinkedAuditsCardFragment\n      }\n    }\n  }\n}\n\nfragment LinkedAuditsCardFragment on Audit {\n  id\n  name\n  createdAt\n  state\n  validFrom\n  validUntil\n  framework {\n    id\n    name\n  }\n}\n"
  }
};
})();

(node as any).hash = "0a9f52be5fb551ec55667b7d777391de";

export default node;
