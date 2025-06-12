/**
 * @generated SignedSource<<ed6783002ee6a5f6c4871f4e3897e1d3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateOrganizationInput = {
  aiFocused?: boolean | null | undefined;
  companyType?: string | null | undefined;
  foundingYear?: number | null | undefined;
  hasEnterpriseAccounts?: boolean | null | undefined;
  hasRaisedMoney?: boolean | null | undefined;
  logo?: any | null | undefined;
  name?: string | null | undefined;
  organizationId: string;
  preMarketFit?: boolean | null | undefined;
  usesAiGeneratedCode?: boolean | null | undefined;
  usesCloudProviders?: boolean | null | undefined;
  vcBacked?: boolean | null | undefined;
};
export type SettingsViewUpdateOrganizationMutation$variables = {
  input: UpdateOrganizationInput;
};
export type SettingsViewUpdateOrganizationMutation$data = {
  readonly updateOrganization: {
    readonly organization: {
      readonly aiFocused: boolean | null | undefined;
      readonly companyType: string | null | undefined;
      readonly foundingYear: number | null | undefined;
      readonly hasEnterpriseAccounts: boolean | null | undefined;
      readonly hasRaisedMoney: boolean | null | undefined;
      readonly id: string;
      readonly logoUrl: string | null | undefined;
      readonly name: string;
      readonly preMarketFit: boolean | null | undefined;
      readonly usesAiGeneratedCode: boolean | null | undefined;
      readonly usesCloudProviders: boolean | null | undefined;
      readonly vcBacked: boolean | null | undefined;
    };
  };
};
export type SettingsViewUpdateOrganizationMutation = {
  response: SettingsViewUpdateOrganizationMutation$data;
  variables: SettingsViewUpdateOrganizationMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "UpdateOrganizationPayload",
    "kind": "LinkedField",
    "name": "updateOrganization",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Organization",
        "kind": "LinkedField",
        "name": "organization",
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
            "name": "name",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "logoUrl",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "foundingYear",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "companyType",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "preMarketFit",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "usesCloudProviders",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "aiFocused",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "usesAiGeneratedCode",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "vcBacked",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "hasRaisedMoney",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "hasEnterpriseAccounts",
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "SettingsViewUpdateOrganizationMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "SettingsViewUpdateOrganizationMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "6aeb228c4f20258fb5287c82314b3f21",
    "id": null,
    "metadata": {},
    "name": "SettingsViewUpdateOrganizationMutation",
    "operationKind": "mutation",
    "text": "mutation SettingsViewUpdateOrganizationMutation(\n  $input: UpdateOrganizationInput!\n) {\n  updateOrganization(input: $input) {\n    organization {\n      id\n      name\n      logoUrl\n      foundingYear\n      companyType\n      preMarketFit\n      usesCloudProviders\n      aiFocused\n      usesAiGeneratedCode\n      vcBacked\n      hasRaisedMoney\n      hasEnterpriseAccounts\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "c668eb7004839c700a07048e2911ce0c";

export default node;
