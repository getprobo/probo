// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import type { INodeProperties } from 'n8n-workflow';
import * as createOp from './create.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as updateOp from './update.operation';
import * as deleteOp from './delete.operation';
import * as createScopeOp from './createScope.operation';
import * as getScopeOp from './getScope.operation';
import * as getAllScopesOp from './getAllScopes.operation';
import * as updateScopeOp from './updateScope.operation';
import * as deleteScopeOp from './deleteScope.operation';
import * as getScopeMermaidChartOp from './getScopeMermaidChart.operation';
import * as createNodeOp from './createNode.operation';
import * as getNodeOp from './getNode.operation';
import * as getAllNodesOp from './getAllNodes.operation';
import * as updateNodeOp from './updateNode.operation';
import * as deleteNodeOp from './deleteNode.operation';
import * as createProcessOp from './createProcess.operation';
import * as getProcessOp from './getProcess.operation';
import * as getAllProcessesOp from './getAllProcesses.operation';
import * as updateProcessOp from './updateProcess.operation';
import * as deleteProcessOp from './deleteProcess.operation';
import * as createThreatOp from './createThreat.operation';
import * as getThreatOp from './getThreat.operation';
import * as getAllThreatsOp from './getAllThreats.operation';
import * as updateThreatOp from './updateThreat.operation';
import * as deleteThreatOp from './deleteThreat.operation';
import * as createScenarioOp from './createScenario.operation';
import * as getScenarioOp from './getScenario.operation';
import * as getAllScenariosOp from './getAllScenarios.operation';
import * as updateScenarioOp from './updateScenario.operation';
import * as deleteScenarioOp from './deleteScenario.operation';
import * as linkScenarioThreatOp from './linkScenarioThreat.operation';
import * as unlinkScenarioThreatOp from './unlinkScenarioThreat.operation';
import * as linkScenarioRiskOp from './linkScenarioRisk.operation';
import * as unlinkScenarioRiskOp from './unlinkScenarioRisk.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['riskAssessment'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a risk assessment',
				action: 'Create a risk assessment',
			},
			{
				name: 'Create Node',
				value: 'createNode',
				description: 'Create a node in a scope',
				action: 'Create a node',
			},
			{
				name: 'Create Process',
				value: 'createProcess',
				description: 'Create a process in a scope',
				action: 'Create a process',
			},
			{
				name: 'Create Scenario',
				value: 'createScenario',
				description: 'Create a scenario in a scope',
				action: 'Create a scenario',
			},
			{
				name: 'Create Scope',
				value: 'createScope',
				description: 'Create a scope in a risk assessment',
				action: 'Create a scope',
			},
			{
				name: 'Create Threat',
				value: 'createThreat',
				description: 'Create a threat in a scope',
				action: 'Create a threat',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a risk assessment',
				action: 'Delete a risk assessment',
			},
			{
				name: 'Delete Node',
				value: 'deleteNode',
				description: 'Delete a node',
				action: 'Delete a node',
			},
			{
				name: 'Delete Process',
				value: 'deleteProcess',
				description: 'Delete a process',
				action: 'Delete a process',
			},
			{
				name: 'Delete Scenario',
				value: 'deleteScenario',
				description: 'Delete a scenario',
				action: 'Delete a scenario',
			},
			{
				name: 'Delete Scope',
				value: 'deleteScope',
				description: 'Delete a scope',
				action: 'Delete a scope',
			},
			{
				name: 'Delete Threat',
				value: 'deleteThreat',
				description: 'Delete a threat',
				action: 'Delete a threat',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a risk assessment',
				action: 'Get a risk assessment',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many risk assessments',
				action: 'Get many risk assessments',
			},
			{
				name: 'Get Many Nodes',
				value: 'getAllNodes',
				action: 'Get many nodes',
			},
			{
				name: 'Get Many Processes',
				value: 'getAllProcesses',
				action: 'Get many processes',
			},
			{
				name: 'Get Many Scenarios',
				value: 'getAllScenarios',
				action: 'Get many scenarios',
			},
			{
				name: 'Get Many Scopes',
				value: 'getAllScopes',
				action: 'Get many scopes',
			},
			{
				name: 'Get Many Threats',
				value: 'getAllThreats',
				action: 'Get many threats',
			},
			{
				name: 'Get Node',
				value: 'getNode',
				description: 'Get a node',
				action: 'Get a node',
			},
			{
				name: 'Get Process',
				value: 'getProcess',
				description: 'Get a process',
				action: 'Get a process',
			},
			{
				name: 'Get Scenario',
				value: 'getScenario',
				description: 'Get a scenario',
				action: 'Get a scenario',
			},
			{
				name: 'Get Scope',
				value: 'getScope',
				description: 'Get a scope',
				action: 'Get a scope',
			},
			{
				name: 'Get Scope Mermaid Chart',
				value: 'getScopeMermaidChart',
				description: 'Get the Mermaid diagram for a scope',
				action: 'Get a scope mermaid chart',
			},
			{
				name: 'Get Threat',
				value: 'getThreat',
				description: 'Get a threat',
				action: 'Get a threat',
			},
			{
				name: 'Link Scenario Risk',
				value: 'linkScenarioRisk',
				description: 'Link a scenario to a risk',
				action: 'Link a scenario to a risk',
			},
			{
				name: 'Link Scenario Threat',
				value: 'linkScenarioThreat',
				description: 'Link a scenario to a threat',
				action: 'Link a scenario to a threat',
			},
			{
				name: 'Unlink Scenario Risk',
				value: 'unlinkScenarioRisk',
				description: 'Unlink a scenario from a risk',
				action: 'Unlink a scenario from a risk',
			},
			{
				name: 'Unlink Scenario Threat',
				value: 'unlinkScenarioThreat',
				description: 'Unlink a scenario from a threat',
				action: 'Unlink a scenario from a threat',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update a risk assessment',
				action: 'Update a risk assessment',
			},
			{
				name: 'Update Node',
				value: 'updateNode',
				description: 'Update a node',
				action: 'Update a node',
			},
			{
				name: 'Update Process',
				value: 'updateProcess',
				description: 'Update a process',
				action: 'Update a process',
			},
			{
				name: 'Update Scenario',
				value: 'updateScenario',
				description: 'Update a scenario',
				action: 'Update a scenario',
			},
			{
				name: 'Update Scope',
				value: 'updateScope',
				description: 'Update a scope',
				action: 'Update a scope',
			},
			{
				name: 'Update Threat',
				value: 'updateThreat',
				description: 'Update a threat',
				action: 'Update a threat',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...getOp.description,
	...getAllOp.description,
	...updateOp.description,
	...deleteOp.description,
	...createScopeOp.description,
	...getScopeOp.description,
	...getAllScopesOp.description,
	...updateScopeOp.description,
	...deleteScopeOp.description,
	...getScopeMermaidChartOp.description,
	...createNodeOp.description,
	...getNodeOp.description,
	...getAllNodesOp.description,
	...updateNodeOp.description,
	...deleteNodeOp.description,
	...createProcessOp.description,
	...getProcessOp.description,
	...getAllProcessesOp.description,
	...updateProcessOp.description,
	...deleteProcessOp.description,
	...createThreatOp.description,
	...getThreatOp.description,
	...getAllThreatsOp.description,
	...updateThreatOp.description,
	...deleteThreatOp.description,
	...createScenarioOp.description,
	...getScenarioOp.description,
	...getAllScenariosOp.description,
	...updateScenarioOp.description,
	...deleteScenarioOp.description,
	...linkScenarioThreatOp.description,
	...unlinkScenarioThreatOp.description,
	...linkScenarioRiskOp.description,
	...unlinkScenarioRiskOp.description,
];

export {
	createOp as create,
	getOp as get,
	getAllOp as getAll,
	updateOp as update,
	deleteOp as delete,
	createScopeOp as createScope,
	getScopeOp as getScope,
	getAllScopesOp as getAllScopes,
	updateScopeOp as updateScope,
	deleteScopeOp as deleteScope,
	getScopeMermaidChartOp as getScopeMermaidChart,
	createNodeOp as createNode,
	getNodeOp as getNode,
	getAllNodesOp as getAllNodes,
	updateNodeOp as updateNode,
	deleteNodeOp as deleteNode,
	createProcessOp as createProcess,
	getProcessOp as getProcess,
	getAllProcessesOp as getAllProcesses,
	updateProcessOp as updateProcess,
	deleteProcessOp as deleteProcess,
	createThreatOp as createThreat,
	getThreatOp as getThreat,
	getAllThreatsOp as getAllThreats,
	updateThreatOp as updateThreat,
	deleteThreatOp as deleteThreat,
	createScenarioOp as createScenario,
	getScenarioOp as getScenario,
	getAllScenariosOp as getAllScenarios,
	updateScenarioOp as updateScenario,
	deleteScenarioOp as deleteScenario,
	linkScenarioThreatOp as linkScenarioThreat,
	unlinkScenarioThreatOp as unlinkScenarioThreat,
	linkScenarioRiskOp as linkScenarioRisk,
	unlinkScenarioRiskOp as unlinkScenarioRisk,
};
