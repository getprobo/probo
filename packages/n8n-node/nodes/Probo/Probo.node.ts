import {
	NodeConnectionTypes,
	type IExecuteFunctions,
	type INodeExecutionData,
	type INodeType,
	type INodeTypeDescription,
} from 'n8n-workflow';
import { executeOperation } from './operations';

export class Probo implements INodeType {
	description: INodeTypeDescription = {
		displayName: 'Probo',
		name: 'probo',
		icon: { light: 'file:../../icons/probo.svg', dark: 'file:../../icons/probo.svg' },
		group: ['input'],
		version: 1,
		subtitle: '={{$parameter["operation"]}}',
		description: 'Consume data from the Probo API',
		defaults: {
			name: 'Probo',
		},
		usableAsTool: true,
		inputs: [NodeConnectionTypes.Main],
		outputs: [NodeConnectionTypes.Main],
		credentials: [
			{
				name: 'proboApi',
				required: true,
				displayOptions: {
					show: {
						authentication: ['apiKey'],
					},
				},
			},
		],
		requestDefaults: {
			headers: {
				Accept: 'application/json',
				'Content-Type': 'application/json',
			},
		},
		properties: [
			{
				displayName: 'Authentication',
				name: 'authentication',
				type: 'options',
				options: [
					{
						name: 'API Key',
						value: 'apiKey',
					},
				],
				default: 'apiKey',
			},
			{
				displayName: 'Operation',
				name: 'operation',
				type: 'options',
				noDataExpression: true,
				options: [
					{
						name: 'Execute',
						value: 'execute',
						description: 'Execute a GraphQL query or mutation',
						action: 'Execute a GraphQL operation',
					},
				],
				default: 'execute',
			},
			{
				displayName: 'Query',
				name: 'query',
				type: 'string',
				typeOptions: {
					rows: 5,
				},
				displayOptions: {
					show: {
						operation: ['execute'],
					},
				},
				default: '',
				description: 'The complete GraphQL operation including operation name and variable declarations (e.g., "query GetUser($userId: ID!) { node(id: $userId) { id } }" or "mutation UpdateUser($input: UpdateUserInput!) { updateUser(input: $input) { id } }")',
				required: true,
			},
			{
				displayName: 'Variables',
				name: 'variables',
				type: 'json',
				displayOptions: {
					show: {
						operation: ['execute'],
					},
				},
				default: '{}',
				description: 'GraphQL variables as JSON object',
			},
		],
	};

	async execute(this: IExecuteFunctions): Promise<INodeExecutionData[][]> {
		const items = this.getInputData();
		const returnData: INodeExecutionData[] = [];

		for (let i = 0; i < items.length; i++) {
			const operation = this.getNodeParameter('operation', i) as string;
			const result = await executeOperation.call(this, operation, i);
			returnData.push(result);
		}

		return [returnData];
	}

	methods = {
		listSearch: {},
	};
}